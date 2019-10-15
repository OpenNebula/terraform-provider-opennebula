package api

import (
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Context struct {
}

type Client struct {
	HTTPClient     *http.Client
	context        *Context
	Authly         *Authly
	Url            string
	progname       string
	currentVersion string
	versionHeader  string
}

func NewClient(c VerticeApi, path string) *Client {
	auth := NewAuthly(c)
	auth.UrlSuffix = path
	return &Client{
		HTTPClient:     &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}},
		context:        &Context{},
		Authly:         auth,
		Url:            auth.GetURL(),
		progname:       "Vertice-Go-api",
		currentVersion: "2",
		versionHeader:  "Supported-Gulp",
	}
}

func (c *Client) detectClientError(err error) error {
	return fmt.Errorf("Failed to connect to api server.")
}

func (c *Client) Do(request *http.Request) (*http.Response, error) {
	for headerKey, headerVal := range c.Authly.AuthMap {
		request.Header.Add(headerKey, headerVal)
	}
	request.Close = true
	response, err := c.HTTPClient.Do(request)

	if err != nil {
		return nil, err
	}
	supported := response.Header.Get(c.versionHeader)
	format := `################################################################

WARNING: You're using an unsupported version of %s.

You must have at least version %s, your current
version is %s.

################################################################

`
	if !validateVersion(supported, c.currentVersion) {
		log.Debugf(format)
		log.Debugf(supported)
	}
	if response.StatusCode > 399 {
		defer response.Body.Close()
		result, _ := ioutil.ReadAll(response.Body)
		return response, errors.New(string(result))
	}
	//defer response.Body.Close()
	return response, nil

}

// validateVersion checks whether current version is greater or equal to
// supported version.
func validateVersion(supported, current string) bool {
	var (
		bigger bool
		limit  int
	)
	if supported == "" {
		return true
	}
	partsSupported := strings.Split(supported, ".")
	partsCurrent := strings.Split(current, ".")
	if len(partsSupported) > len(partsCurrent) {
		limit = len(partsCurrent)
		bigger = true
	} else {
		limit = len(partsSupported)
	}
	for i := 0; i < limit; i++ {
		current, err := strconv.Atoi(partsCurrent[i])
		if err != nil {
			return false
		}
		supported, err := strconv.Atoi(partsSupported[i])
		if err != nil {
			return false
		}
		if current < supported {
			return false
		}
		if current > supported {
			return true
		}
	}
	if bigger {
		return false
	}
	return true
}
