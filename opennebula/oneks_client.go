package opennebula

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const oneKSRequestTimeout = time.Minute

type oneKSBody struct {
	Name              string           `json:"name"`
	Description       string           `json:"description"`
	State             string           `json:"state"`
	KubernetesVersion string           `json:"kubernetes_version"`
	Deployment        *oneKSDeployment `json:"deployment"`
	ControlPlane      *oneKSSpec       `json:"control_plane"`
}

type oneKSReference struct {
	ID int `json:"id"`
}

func (r *oneKSReference) UnmarshalJSON(data []byte) error {
	var ref struct {
		ID *int `json:"id"`
	}
	if err := json.Unmarshal(data, &ref); err != nil {
		return err
	}
	if ref.ID == nil || *ref.ID < 0 {
		return fmt.Errorf("missing or invalid OneKS placement ID")
	}
	r.ID = *ref.ID
	return nil
}

type oneKSDeployment struct {
	Cluster    *oneKSReference  `json:"cluster"`
	Networks   *oneKSNetworks   `json:"networks"`
	Datastores *oneKSDatastores `json:"datastores,omitempty"`
}

type oneKSNetworks struct {
	Public  *oneKSReference `json:"public"`
	Private *oneKSReference `json:"private"`
}

type oneKSDatastores struct {
	Image *oneKSReference `json:"image,omitempty"`
}

type oneKSSpec struct {
	Family     string          `json:"family,omitempty"`
	Flavour    string          `json:"flavour"`
	UserInputs json.RawMessage `json:"user_inputs_values,omitempty"`
}

type oneKSDocument struct {
	ID       json.Number `json:"ID"`
	Template struct {
		Body *oneKSBody `json:"CLUSTER_BODY"`
	} `json:"TEMPLATE"`
}

type oneKSResponse struct {
	Document *oneKSDocument `json:"DOCUMENT"`
}

func (r *oneKSResponse) validate() error {
	if r == nil || r.Document == nil {
		return fmt.Errorf("invalid OneKS response: missing DOCUMENT")
	}
	id, err := strconv.Atoi(r.Document.ID.String())
	if err != nil || id < 0 {
		return fmt.Errorf("invalid OneKS response: missing or invalid cluster ID")
	}
	body := r.Document.Template.Body
	if body == nil || body.Name == "" || body.State == "" || body.KubernetesVersion == "" || body.Deployment == nil {
		return fmt.Errorf("invalid OneKS cluster %d: missing or incomplete CLUSTER_BODY", id)
	}
	deployment := body.Deployment
	if deployment.Cluster == nil || deployment.Networks == nil || deployment.Networks.Public == nil || deployment.Networks.Private == nil {
		return fmt.Errorf("invalid OneKS cluster %d: incomplete deployment", id)
	}
	return nil
}

func oneKSJSONObject(value []byte) bool {
	var object map[string]json.RawMessage
	return json.Unmarshal(value, &object) == nil && object != nil
}

type oneKSAPIError struct {
	Status int
	Method string
	Path   string
}

func (e *oneKSAPIError) Error() string {
	return fmt.Sprintf("OneKS API %s %s returned HTTP %d (%s)", e.Method, e.Path, e.Status, http.StatusText(e.Status))
}

func oneKSNotFound(err error) bool {
	var apiErr *oneKSAPIError
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound
}

type oneKSClient struct {
	endpoint string
	username string
	password string
	http     *http.Client
}

func validateOneKSEndpoint(v interface{}, k string) ([]string, []error) {
	u, err := url.Parse(v.(string))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, []error{fmt.Errorf("%s must be an HTTP(S) API base URL without credentials, query, or fragment", k)}
	}
	return nil, nil
}

func newOneKSClient(endpoint, username, password string, client *http.Client) (*oneKSClient, error) {
	if _, errs := validateOneKSEndpoint(endpoint, "oneks_endpoint"); len(errs) > 0 {
		return nil, errs[0]
	}
	httpClient := *client
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	return &oneKSClient{
		endpoint: strings.TrimRight(endpoint, "/"),
		username: username,
		password: password,
		http:     &httpClient,
	}, nil
}

func getOneKSClient(meta interface{}) (*oneKSClient, error) {
	config := meta.(*Configuration)
	if config.OneKS == nil {
		return nil, fmt.Errorf("oneks_endpoint (or OPENNEBULA_ONEKS_ENDPOINT) must be configured to manage OneKS clusters")
	}
	return config.OneKS, nil
}

func (c *oneKSClient) request(ctx context.Context, method, path string, input, output interface{}) error {
	var payload bytes.Buffer
	if input != nil {
		if err := json.NewEncoder(&payload).Encode(input); err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, &payload)
	if err != nil {
		return fmt.Errorf("failed to build OneKS API request")
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return fmt.Errorf("OneKS API %s %s request failed: %w", method, path, err)
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return &oneKSAPIError{
			Status: res.StatusCode,
			Method: method,
			Path:   path,
		}
	}
	if output == nil {
		return nil
	}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(output); err != nil {
		return fmt.Errorf("invalid JSON response from OneKS API %s %s", method, path)
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("invalid trailing data from OneKS API %s %s", method, path)
	}
	return nil
}

func (c *oneKSClient) cluster(ctx context.Context, id string) (*oneKSResponse, error) {
	return c.clusterWithQuery(ctx, id, "")
}

func (c *oneKSClient) clusterWithQuery(ctx context.Context, id, query string) (*oneKSResponse, error) {
	if n, err := strconv.Atoi(id); err != nil || n < 0 {
		return nil, fmt.Errorf("invalid OneKS cluster ID %q", id)
	}
	var result oneKSResponse
	if err := c.request(ctx, http.MethodGet, "/clusters/"+id+query, nil, &result); err != nil {
		return nil, err
	}
	if err := result.validate(); err != nil {
		return nil, err
	}
	if result.Document.ID.String() != id {
		return nil, fmt.Errorf("OneKS API returned a different cluster ID")
	}
	return &result, nil
}
