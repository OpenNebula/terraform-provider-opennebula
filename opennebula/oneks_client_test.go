package opennebula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestOneKSClientRequest(t *testing.T) {
	config := testOneKSConfig(t, func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "user" || password != "secret" {
			t.Error("missing basic authentication")
		}
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/clusters" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		for _, header := range []string{"Content-Type", "Accept"} {
			if got := r.Header.Get(header); got != "application/json" {
				t.Errorf("%s = %q, want application/json", header, got)
			}
		}
		fmt.Fprint(w, `{"ok":true}`)
	})
	var result map[string]bool
	if err := config.OneKS.request(context.Background(), http.MethodPost, "/clusters", map[string]string{"name": "test"}, &result); err != nil {
		t.Fatal(err)
	}
	if !result["ok"] {
		t.Fatalf("unexpected response: %v", result)
	}
}

func TestOneKSClientRejectsRedirect(t *testing.T) {
	config := testOneKSConfig(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/api/v1/clusters/other")
		w.WriteHeader(http.StatusFound)
	})
	if _, err := config.OneKS.cluster(context.Background(), testOneKSClusterID); err == nil || !strings.Contains(err.Error(), "302") {
		t.Fatalf("API redirect should fail: %v", err)
	}
}

func TestOneKSInvalidIDsAndEndpoints(t *testing.T) {
	for _, endpoint := range []string{
		"",
		"://bad",
		"file:///tmp/oneks",
		"http://user:secret@example.com/api/v1",
		"http://example.com?token=secret",
		"http://example.com#fragment",
	} {
		if _, errs := validateOneKSEndpoint(endpoint, "oneks_endpoint"); len(errs) == 0 {
			t.Errorf("invalid endpoint accepted: %s", endpoint)
		}
	}
	for _, endpoint := range []string{"http://example.com:10780/api/v1", "https://example.com/oneks/api/v1/"} {
		if _, errs := validateOneKSEndpoint(endpoint, "oneks_endpoint"); len(errs) != 0 {
			t.Errorf("valid endpoint rejected: %s", endpoint)
		}
	}
	config := testOneKSConfig(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid cluster ID reached the API") })
	for _, id := range []string{"", "-1", "1/other", "bad"} {
		if _, err := config.OneKS.cluster(context.Background(), id); err == nil {
			t.Errorf("invalid ID accepted: %s", id)
		}
	}
}

// Placement IDs must be present and decode as nonnegative integers.
func TestOneKSInvalidPlacementID(t *testing.T) {
	for _, body := range []string{`{}`, `{"id":-1}`, `{"id":1.5}`, `{"id":"0"}`} {
		var ref oneKSReference
		if err := json.Unmarshal([]byte(body), &ref); err == nil {
			t.Errorf("accepted invalid placement: %s", body)
		}
	}
}
