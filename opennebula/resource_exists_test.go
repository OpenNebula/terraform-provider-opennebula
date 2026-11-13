package opennebula

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	gocaerrors "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type existsTestTransport func(*http.Request) (*http.Response, error)

func (f existsTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type existsTestResponse struct {
	name       string
	body       string
	status     int
	transport  error
	apiCode    gocaerrors.OneErrCode
	clientCode gocaerrors.ClientErrCode
	missing    bool
}

func existsTestResponses() []existsTestResponse {
	apiFailure := func(name string, code gocaerrors.OneErrCode) existsTestResponse {
		return existsTestResponse{
			name: name,
			body: fmt.Sprintf(`<?xml version="1.0"?>
<methodResponse><params><param><value><array><data>
<value><boolean>0</boolean></value>
<value><string>mock API failure</string></value>
<value><int>%d</int></value>
</data></array></value></param></params></methodResponse>`, code),
			apiCode: code,
			missing: code == gocaerrors.OneNoExistsError,
		}
	}
	return []existsTestResponse{
		apiFailure("not_found", gocaerrors.OneNoExistsError),
		apiFailure("authentication", gocaerrors.OneAuthenticationError),
		apiFailure("authorization", gocaerrors.OneAuthorizationError),
		apiFailure("action", gocaerrors.OneActionError),
		{
			name:       "transport",
			transport:  errors.New("mock connection failure"),
			clientCode: gocaerrors.ClientReqHTTP,
		},
		{
			name:       "http_500",
			status:     http.StatusInternalServerError,
			body:       "mock server failure",
			clientCode: gocaerrors.ClientRespHTTP,
		},
		{
			name:       "malformed_xml",
			body:       "not XML",
			clientCode: gocaerrors.ClientRespXMLRPCParse,
		},
	}
}

func newExistsTestConfiguration(t *testing.T, method string, response existsTestResponse) (*Configuration, *int) {
	t.Helper()
	calls := 0
	transport := existsTestTransport(func(req *http.Request) (*http.Response, error) {
		calls++
		var request struct {
			Method string `xml:"methodName"`
		}
		if err := xml.NewDecoder(req.Body).Decode(&request); err != nil {
			t.Errorf("decode XML-RPC request: %v", err)
		}
		if req.Method != http.MethodPost || request.Method != method {
			t.Errorf("request = %s %q, want POST %q", req.Method, request.Method, method)
		}
		if response.transport != nil {
			return nil, response.transport
		}
		status := response.status
		if status == 0 {
			status = http.StatusOK
		}
		return &http.Response{
			StatusCode: status,
			Header:     http.Header{"Content-Type": []string{"text/xml"}},
			Body:       io.NopCloser(strings.NewReader(response.body)),
			Request:    req,
		}, nil
	})
	client := goca.NewClient(goca.OneConfig{
		Token:    "test:test",
		Endpoint: "http://opennebula.invalid/RPC2",
	}, &http.Client{Transport: transport})
	return &Configuration{Controller: goca.NewController(client)}, &calls
}

func TestXMLRPCResourceExists(t *testing.T) {
	resources := []struct {
		name     string
		resource func() *schema.Resource
		method   string
		parent   string
	}{
		{"image", resourceOpennebulaImage, "one.image.info", ""},
		{"security_group", resourceOpennebulaSecurityGroup, "one.secgroup.info", ""},
		{"template", resourceOpennebulaTemplate, "one.template.info", ""},
		{"virtual_machine", resourceOpennebulaVirtualMachine, "one.vm.info", ""},
		{"virtual_network", resourceOpennebulaVirtualNetwork, "one.vn.info", ""},
		{"virtual_router", resourceOpennebulaVirtualRouter, "one.vrouter.info", ""},
		{"vm_group", resourceOpennebulaVMGroup, "one.vmgroup.info", ""},
		{"virtual_router_instance", resourceOpennebulaVirtualRouterInstance, "one.vm.info", ""},
		{"virtual_router_instance_template", resourceOpennebulaVirtualRouterInstanceTemplate, "one.template.info", ""},
		{"virtual_network_address_range", resourceOpennebulaVirtualNetworkAddressRange, "one.vn.info", "virtual_network_id"},
		{"virtual_router_nic", resourceOpennebulaVirtualRouterNIC, "one.vrouter.info", "virtual_router_id"},
	}
	for _, resource := range resources {
		t.Run(resource.name, func(t *testing.T) {
			for _, response := range existsTestResponses() {
				t.Run(response.name, func(t *testing.T) {
					r := resource.resource()
					values := map[string]interface{}{}
					if resource.parent != "" {
						values[resource.parent] = 42
					}
					d := schema.TestResourceDataRaw(t, r.Schema, values)
					d.SetId("42")
					config, calls := newExistsTestConfiguration(t, resource.method, response)
					exists, err := r.Exists(d, config)
					if response.missing {
						if exists || err != nil {
							t.Fatalf("Exists = (%v, %v), want (false, nil)", exists, err)
						}
					} else {
						if exists || err == nil {
							t.Fatalf("Exists = (%v, %v), want (false, error)", exists, err)
						}
						if response.apiCode != 0 {
							var apiErr *gocaerrors.ResponseError
							if !errors.As(err, &apiErr) || apiErr.Code != response.apiCode || apiErr.Msg != "mock API failure" {
								t.Errorf("API error not preserved: %v", err)
							}
						} else {
							var clientErr *gocaerrors.ClientError
							if !errors.As(err, &clientErr) || clientErr.Code != response.clientCode {
								t.Errorf("client error not preserved: %v", err)
							}
						}
					}
					if *calls != 1 {
						t.Errorf("XML-RPC calls = %d, want 1", *calls)
					}
					if d.Id() != "42" {
						t.Errorf("Exists changed state ID to %q", d.Id())
					}
				})
			}
			if resource.parent == "" {
				t.Run("invalid_id", func(t *testing.T) {
					r := resource.resource()
					d := schema.TestResourceDataRaw(t, r.Schema, nil)
					d.SetId("invalid")
					config, calls := newExistsTestConfiguration(t, resource.method, existsTestResponse{})
					exists, err := r.Exists(d, config)
					if exists || err == nil || *calls != 0 || d.Id() != "invalid" {
						t.Fatalf("invalid ID: Exists = (%v, %v), calls = %d, ID = %q", exists, err, *calls, d.Id())
					}
				})
			}
		})
	}
}

func TestVirtualNetworkRefreshWithoutUpgradeExists(t *testing.T) {
	for _, response := range existsTestResponses() {
		t.Run(response.name, func(t *testing.T) {
			r := resourceOpennebulaVirtualNetwork()
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "test-network"})
			d.SetId("42")
			original := d.State()
			config, calls := newExistsTestConfiguration(t, "one.vn.info", response)
			state, diags := r.RefreshWithoutUpgrade(context.Background(), original, config)
			if response.missing {
				if state != nil || diags.HasError() {
					t.Fatalf("missing network: state = %#v, diagnostics = %v; want nil state without errors", state, diags)
				}
			} else {
				if !diags.HasError() {
					t.Fatal("expected refresh error diagnostics")
				}
				if state == nil || state.ID != "42" || state.Attributes["name"] != "test-network" {
					t.Fatalf("refresh discarded state on a real error: %#v", state)
				}
			}
			if *calls != 1 {
				t.Errorf("XML-RPC calls = %d, want 1 (Exists must short-circuit Read)", *calls)
			}
		})
	}
}
