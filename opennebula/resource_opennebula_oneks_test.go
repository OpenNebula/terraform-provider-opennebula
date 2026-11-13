package opennebula

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testOneKSClusterID = "1"

const testOneKSInputs = `{
  "count": 0,
  "enabled": false,
  "zones": ["a", "b"],
  "settings": {"id": 3},
  "unset": null
}`

const testOneKSDeployment = `{
  "cluster": {"id": 0},
  "networks": {
    "public": {"id": 1},
    "private": {"id": 2}
  },
  "datastores": {"image": {"id": 0}}
}`

func TestAccOneKS(t *testing.T) {
	if os.Getenv("OPENNEBULA_ONEKS_ENDPOINT") == "" {
		t.Skip("OPENNEBULA_ONEKS_ENDPOINT is required for OneKS acceptance tests")
	}
	name := fmt.Sprintf("tf-oneks-%d", time.Now().UnixNano())
	basic := testAccOneKSConfig(name)
	updated := testAccOneKSConfig(name + "-updated")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testEnvIsSet("OPENNEBULA_ONEKS_KUBERNETES_VERSION", t)
			for _, key := range []string{"OPENNEBULA_ONEKS_CLUSTER_ID", "OPENNEBULA_ONEKS_PUBLIC_NETWORK_ID", "OPENNEBULA_ONEKS_PRIVATE_NETWORK_ID"} {
				id, err := strconv.Atoi(os.Getenv(key))
				if err != nil || id < 0 {
					t.Fatalf("%s must be a nonnegative resource ID", key)
				}
			}
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOneKSDestroy,
		Steps: []resource.TestStep{
			{Config: basic, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("opennebula_oneks.test", "name", name),
				resource.TestCheckResourceAttr("opennebula_oneks.test", "state", "RUNNING"),
				resource.TestCheckResourceAttrPair("data.opennebula_oneks.by_id", "id", "opennebula_oneks.test", "id"),
				resource.TestCheckResourceAttrPair("data.opennebula_oneks.by_name", "id", "opennebula_oneks.test", "id"),
			)},
			{Config: updated, Check: resource.TestCheckResourceAttr("opennebula_oneks.test", "name", name+"-updated")},
			{ResourceName: "opennebula_oneks.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"force_delete"}},
		},
	})
}

func testAccCheckOneKSDestroy(state *terraform.State) error {
	client := testAccProvider.Meta().(*Configuration).OneKS
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "opennebula_oneks" {
			continue
		}
		_, err := client.cluster(context.Background(), rs.Primary.ID)
		if oneKSNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("OneKS cluster %s still exists", rs.Primary.ID)
	}
	return nil
}

func TestOneKSCreateErrors(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		body, id string
	}{
		{
			name:   "request rejected",
			status: http.StatusBadRequest,
			body:   `{"error":"secret"}`,
		},
		{
			name:   "provisioning failed",
			status: http.StatusCreated,
			body:   testOneKSResponse("0", "test-cluster", "PROVISIONING_FAILURE"),
			id:     "0",
		},
		{
			name:   "incomplete response",
			status: http.StatusCreated,
			body:   fmt.Sprintf(`{"DOCUMENT":{"ID":%q,"TEMPLATE":{}}}`, testOneKSClusterID),
			id:     testOneKSClusterID,
		},
		{
			name:   "trailing JSON",
			status: http.StatusCreated,
			body:   testOneKSResponse(testOneKSClusterID, "test-cluster", "RUNNING") + `{}`,
			id:     testOneKSClusterID,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := testOneKSConfig(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			})
			d := testOneKSData(t)
			diags := resourceOpennebulaOneKSCreate(context.Background(), d, config)
			if !diags.HasError() || d.Id() != tc.id || strings.Contains(fmt.Sprint(diags), "secret") {
				t.Fatalf("id = %q, want %q; diagnostics: %v", d.Id(), tc.id, diags)
			}
		})
	}
}

func TestOneKSReadAndDeleteErrors(t *testing.T) {
	for _, operation := range []string{"read", "delete"} {
		for status, wantID := range map[int]string{
			http.StatusUnauthorized:        testOneKSClusterID,
			http.StatusForbidden:           testOneKSClusterID,
			http.StatusNotFound:            "",
			http.StatusInternalServerError: testOneKSClusterID,
		} {
			t.Run(fmt.Sprintf("%s/%d", operation, status), func(t *testing.T) {
				config := testOneKSConfig(t, func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(status)
					fmt.Fprint(w, `{"error":"secret"}`)
				})
				d := testOneKSData(t)
				d.SetId(testOneKSClusterID)
				callback := resourceOpennebulaOneKSRead
				if operation == "delete" {
					callback = resourceOpennebulaOneKSDelete
				}
				diags := callback(context.Background(), d, config)
				if diags.HasError() != (status != http.StatusNotFound) || d.Id() != wantID || strings.Contains(fmt.Sprint(diags), "secret") {
					t.Fatalf("id = %q; diagnostics: %v", d.Id(), diags)
				}
			})
		}
	}
}

func TestOneKSDelete(t *testing.T) {
	for state, wantID := range map[string]string{
		"RUNNING":                "",
		"DEPROVISIONING":         "",
		"DEPROVISIONING_FAILURE": testOneKSClusterID,
	} {
		t.Run(state, func(t *testing.T) {
			var reads atomic.Int32
			var deleted atomic.Bool
			config := testOneKSConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodDelete {
					deleted.Store(true)
					if r.URL.Query().Get("force") != "true" {
						t.Error("missing force flag")
					}
					w.WriteHeader(http.StatusAccepted)
					return
				}
				if reads.Add(1) > 1 && state != "DEPROVISIONING_FAILURE" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				fmt.Fprint(w, testOneKSResponse(testOneKSClusterID, "test-cluster", state))
			})
			d := testOneKSData(t)
			d.SetId(testOneKSClusterID)
			if err := d.Set("force_delete", true); err != nil {
				t.Fatal(err)
			}
			diags := resourceOpennebulaOneKSDelete(context.Background(), d, config)
			failed := wantID != ""
			if diags.HasError() != failed || d.Id() != wantID || deleted.Load() != (state != "DEPROVISIONING") {
				t.Fatalf("id = %q; diagnostics: %v", d.Id(), diags)
			}
		})
	}
}

func TestOneKSWaitDeadline(t *testing.T) {
	config := testOneKSConfig(t, func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() })
	if _, err := waitOneKSCluster(context.Background(), config.OneKS, testOneKSClusterID, true, 50*time.Millisecond); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestOneKSReplacement(t *testing.T) {
	for _, field := range []string{
		"deployment.0.cluster.0.id", "deployment.0.networks.0.public.0.id",
		"deployment.0.networks.0.private.0.id", "deployment.0.datastores.0.image.0.id",
		"spec.0.family", "spec.0.flavour", "kubernetes_version",
	} {
		t.Run(field, func(t *testing.T) {
			r := resourceOpennebulaOneKS()
			d := testOneKSData(t)
			d.SetId(testOneKSClusterID)
			state := d.State()
			state.Attributes[field] = "100"
			diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(testOneKSValues()), nil)
			if err != nil || diff == nil || !diff.RequiresNew() {
				t.Fatalf("expected replacement: %v, %v", diff, err)
			}
		})
	}
}

func TestOneKSJSONDiff(t *testing.T) {
	for _, tc := range []struct {
		name     string
		old, new string
		replace  bool
	}{
		{
			name: "equivalent formatting and numbers",
			old:  `{"count":1,"enabled":false}`,
			new:  `{ "enabled": false, "count": 1.0 }`,
		},
		{
			name:    "changed input value",
			old:     `{"count":1}`,
			new:     `{"count":2}`,
			replace: true,
		},
		{
			name:    "string changed to number",
			old:     `{"count":"1"}`,
			new:     `{"count":1}`,
			replace: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := resourceOpennebulaOneKS()
			d := testOneKSData(t)
			d.SetId(testOneKSClusterID)
			state := d.State()
			state.Attributes["spec.0.user_inputs_values"] = tc.old
			values := testOneKSValues()
			values["spec"].([]interface{})[0].(map[string]interface{})["user_inputs_values"] = tc.new
			diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(values), nil)
			if err != nil {
				t.Fatal(err)
			}
			changed := diff != nil && !diff.Empty()
			replaces := diff != nil && diff.RequiresNew()
			if changed != tc.replace || replaces != tc.replace {
				t.Fatalf("unexpected JSON diff: %v", diff)
			}
		})
	}
}

func testOneKSConfig(t *testing.T, handler http.HandlerFunc) *Configuration {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := newOneKSClient(server.URL+"/api/v1/", "user", "secret", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	return &Configuration{OneKS: client}
}

func testOneKSData(t *testing.T) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, resourceOpennebulaOneKS().Schema, testOneKSValues())
}

func testOneKSValues() map[string]interface{} {
	return map[string]interface{}{
		"name":               "test-cluster",
		"kubernetes_version": "v1.35.0",
		"deployment": []interface{}{map[string]interface{}{
			"cluster": []interface{}{map[string]interface{}{"id": 0}},
			"networks": []interface{}{map[string]interface{}{
				"public":  []interface{}{map[string]interface{}{"id": 1}},
				"private": []interface{}{map[string]interface{}{"id": 2}},
			}},
			"datastores": []interface{}{map[string]interface{}{"image": []interface{}{map[string]interface{}{"id": 0}}}},
		}},
		"spec": []interface{}{map[string]interface{}{
			"family":             "general",
			"flavour":            "standalone",
			"user_inputs_values": testOneKSInputs,
		}},
	}
}

func testOneKSResponse(id, name, state string) string {
	return fmt.Sprintf(`{
  "DOCUMENT": {
    "ID": %q,
    "TEMPLATE": {
      "CLUSTER_BODY": {
        "name": %q,
        "description": "test cluster",
        "state": %q,
        "kubernetes_version": "v1.35.0",
        "deployment": %s,
        "control_plane": null
      }
    }
  }
}`, id, name, state, testOneKSDeployment)
}

func testAccOneKSConfig(name string) string {
	return fmt.Sprintf(`
resource "opennebula_oneks" "test" {
  name = %q
  description = "Terraform acceptance test"
  force_delete = true
  kubernetes_version = %q
  deployment {
    cluster { id = %s }
    networks {
      public { id = %s }
      private { id = %s }
    }
  }
  spec {
    family = "general"
    flavour = "standalone"
    user_inputs_values = jsonencode({
      count = 1
      cpu = 2
      vcpu = 2
      memory = 4096
      disk_size = 16384
    })
  }
}

data "opennebula_oneks" "by_id" {
  id = opennebula_oneks.test.id
}
data "opennebula_oneks" "by_name" {
  name = opennebula_oneks.test.name
}
`, name, os.Getenv("OPENNEBULA_ONEKS_KUBERNETES_VERSION"), os.Getenv("OPENNEBULA_ONEKS_CLUSTER_ID"),
		os.Getenv("OPENNEBULA_ONEKS_PUBLIC_NETWORK_ID"), os.Getenv("OPENNEBULA_ONEKS_PRIVATE_NETWORK_ID"))
}
