package opennebula

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"testing"

	gocaerrors "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestMissingVirtualNetworkPlansRecreation(t *testing.T) {
	r := resourceOpennebulaVirtualNetwork()
	raw := map[string]interface{}{"name": "test-network"}
	d := schema.TestResourceDataRaw(t, r.Schema, raw)
	d.SetId("42")
	cfg, _ := newExistsTestConfiguration(t, "one.vn.info", existsTestResponses()[0])
	state, diags := r.RefreshWithoutUpgrade(context.Background(), d.State(), cfg)
	if diags.HasError() || state != nil {
		t.Fatalf("missing network: state=%#v, diagnostics=%v", state, diags)
	}
	diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(raw), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if diff == nil || diff.Empty() || diff.Destroy {
		t.Fatalf("expected a creation plan, got %#v", diff)
	}
	if name := diff.Attributes["name"]; name == nil || name.New != "test-network" {
		t.Fatalf("creation plan lost configured name: %#v", name)
	}
}

func TestNoExistsWrappedErrors(t *testing.T) {
	missing := &gocaerrors.ResponseError{Code: gocaerrors.OneNoExistsError, Msg: "missing"}
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"missing", missing, true},
		{"wrapped missing", fmt.Errorf("read: %w", missing), true},
		{"authorization", &gocaerrors.ResponseError{Code: gocaerrors.OneAuthorizationError}, false},
		{"wrapped authorization", fmt.Errorf("read: %w", &gocaerrors.ResponseError{Code: gocaerrors.OneAuthorizationError}), false},
		{"untyped message", errors.New("resource does not exist"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := NoExists(tc.err); got != tc.want {
				t.Fatalf("NoExists(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestSubresourceReadMissingParent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		resource func() *schema.Resource
		method   string
		parent   string
	}{
		{"group admins", resourceOpennebulaGroupAdmins, "one.group.info", ""},
		{"group quotas", resourceOpennebulaGroupQuotas, "one.group.info", ""},
		{"user quotas", resourceOpennebulaUserQuotas, "one.user.info", ""},
		{"address range", resourceOpennebulaVirtualNetworkAddressRange, "one.vn.info", "virtual_network_id"},
		{"router nic", resourceOpennebulaVirtualRouterNIC, "one.vrouter.info", "virtual_router_id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, response := range existsTestResponses() {
				t.Run(response.name, func(t *testing.T) {
					r := tc.resource()
					values := map[string]interface{}{}
					if tc.parent != "" {
						values[tc.parent] = 42
					}
					d := schema.TestResourceDataRaw(t, r.Schema, values)
					d.SetId("7")
					cfg, _ := newExistsTestConfiguration(t, tc.method, response)
					diags := r.ReadContext(context.Background(), d, cfg)
					if response.missing {
						if diags.HasError() || d.Id() != "" {
							t.Fatalf("missing parent: ID=%q, diagnostics=%v", d.Id(), diags)
						}
					} else if !diags.HasError() || d.Id() != "7" {
						t.Fatalf("real error must preserve ID: ID=%q, diagnostics=%v", d.Id(), diags)
					}
				})
			}
		})
	}
}

func TestRefreshMissingListResource(t *testing.T) {
	for _, tc := range []struct {
		name     string
		resource func() *schema.Resource
		method   string
		parent   string
		info     string
	}{
		{"acl", resourceOpennebulaACL, "one.acl.info", "", `<ACL_POOL/>`},
		{"address range", resourceOpennebulaVirtualNetworkAddressRange, "one.vn.info", "virtual_network_id", `<VNET><ID>42</ID><AR_POOL/></VNET>`},
		{"router nic", resourceOpennebulaVirtualRouterNIC, "one.vrouter.info", "virtual_router_id", `<VROUTER><ID>42</ID><TEMPLATE/></VROUTER>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.resource()
			values := map[string]interface{}{}
			if tc.parent != "" {
				values[tc.parent] = 42
			}
			d := schema.TestResourceDataRaw(t, r.Schema, values)
			d.SetId("7")
			var escaped strings.Builder
			if err := xml.EscapeText(&escaped, []byte(tc.info)); err != nil {
				t.Fatal(err)
			}
			body := `<methodResponse><params><param><value><array><data><value><boolean>1</boolean></value><value><string>` + escaped.String() + `</string></value><value><int>0</int></value></data></array></value></param></params></methodResponse>`
			cfg, _ := newExistsTestConfiguration(t, tc.method, existsTestResponse{body: body})
			state, diags := r.RefreshWithoutUpgrade(context.Background(), d.State(), cfg)
			if diags.HasError() || state != nil {
				t.Fatalf("missing list entry: state=%#v, diagnostics=%v", state, diags)
			}
		})
	}
}
