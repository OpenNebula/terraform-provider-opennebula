package opennebula

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestVirtualNetworkSecurityGroupsUpdate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		groups []interface{}
	}{
		{name: "replace", groups: []interface{}{100}},
		{name: "multiple", groups: []interface{}{100, 200}},
		{name: "clear", groups: []interface{}{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			updates := 0
			backendGroups := "0"
			transport := vnetTypeUpdateTransport(func(req *http.Request) (*http.Response, error) {
				var call struct {
					Method string `xml:"methodName"`
					Params []struct {
						Value struct {
							String string `xml:"string"`
						} `xml:"value"`
					} `xml:"params>param"`
				}
				if err := xml.NewDecoder(req.Body).Decode(&call); err != nil {
					return nil, err
				}
				var value string
				switch call.Method {
				case "one.vn.info":
					info := `<VNET><ID>9</ID><NAME>test-vnet</NAME><VN_MAD>bridge</VN_MAD><PERMISSIONS/><CLUSTERS/><TEMPLATE><MTU>1500</MTU><GUEST_MTU>1500</GUEST_MTU><CUSTOM_TAG>keep</CUSTOM_TAG>`
					if backendGroups != "" {
						info += `<SECURITY_GROUPS>` + backendGroups + `</SECURITY_GROUPS>`
					}
					info += `</TEMPLATE><AR_POOL/></VNET>`
					var escaped bytes.Buffer
					if err := xml.EscapeText(&escaped, []byte(info)); err != nil {
						return nil, err
					}
					value = "<string>" + escaped.String() + "</string>"
				case "one.vn.update":
					updates++
					sent := call.Params[2].Value.String
					if !strings.Contains(sent, `CUSTOM_TAG="keep"`) {
						t.Errorf("existing attribute lost: %s", sent)
					}
					backendGroups = ""
					if _, rest, ok := strings.Cut(sent, `SECURITY_GROUPS="`); ok {
						backendGroups, _, _ = strings.Cut(rest, `"`)
					}
					actual := []interface{}{}
					if backendGroups != "" {
						for _, entry := range strings.Split(backendGroups, ",") {
							id, err := strconv.Atoi(entry)
							if err != nil {
								return nil, err
							}
							actual = append(actual, id)
						}
					}
					if !schema.NewSet(schema.HashInt, actual).Equal(schema.NewSet(schema.HashInt, tc.groups)) {
						t.Errorf("sent security groups %v, want %v", actual, tc.groups)
					}
					value = "<int>9</int>"
				default:
					return nil, fmt.Errorf("unexpected XML-RPC method %s", call.Method)
				}
				body := `<methodResponse><params><param><value><array><data><value><boolean>1</boolean></value><value>` + value + `</value><value><int>0</int></value></data></array></value></param></params></methodResponse>`
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			cfg := &Configuration{Controller: goca.NewController(goca.NewClient(goca.NewConfig("test", "test", "http://one.example/RPC2"), &http.Client{Transport: transport}))}
			r := resourceOpennebulaVirtualNetwork()
			old := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "test-vnet", "security_groups": []interface{}{0}})
			old.SetId("9")
			raw := map[string]interface{}{"name": "test-vnet", "security_groups": tc.groups}
			diff, err := r.Diff(context.Background(), old.State(), terraform.NewResourceConfigRaw(raw), cfg)
			if err != nil {
				t.Fatal(err)
			}
			// Isolate the security-group change from computed fields absent in the fixture.
			for key := range diff.Attributes {
				if !strings.HasPrefix(key, "security_groups.") {
					delete(diff.Attributes, key)
				}
			}
			state, diags := r.Apply(context.Background(), old.State(), diff, cfg)
			if diags.HasError() {
				t.Fatal(diags)
			}
			if updates != 1 {
				t.Fatalf("got %d updates, want 1", updates)
			}
			if got := state.Attributes["security_groups.#"]; got != strconv.Itoa(len(tc.groups)) {
				t.Errorf("security group count after refresh = %s, want %d", got, len(tc.groups))
			}
		})
	}
}
