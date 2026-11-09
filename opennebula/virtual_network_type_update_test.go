package opennebula

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type vnetTypeUpdateTransport func(*http.Request) (*http.Response, error)

func (f vnetTypeUpdateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestVirtualNetworkTypeOnlyUpdate(t *testing.T) {
	for _, reject := range []bool{false, true} {
		t.Run(fmt.Sprintf("reject=%t", reject), func(t *testing.T) {
			updates := 0
			backendType := "bridge"
			transport := vnetTypeUpdateTransport(func(req *http.Request) (*http.Response, error) {
				var call struct {
					Method string `xml:"methodName"`
					Params []struct {
						Value struct {
							String string `xml:"string"`
							Int    int    `xml:"int"`
						} `xml:"value"`
					} `xml:"params>param"`
				}
				if err := xml.NewDecoder(req.Body).Decode(&call); err != nil {
					return nil, err
				}
				var value string
				success := "1"
				switch call.Method {
				case "one.vn.info":
					info := `<VNET><ID>9</ID><NAME>test-vnet</NAME><UID>0</UID><GID>0</GID><UNAME>oneadmin</UNAME><GNAME>oneadmin</GNAME><BRIDGE>brtest</BRIDGE><VN_MAD>` + backendType + `</VN_MAD><PERMISSIONS><OWNER_U>1</OWNER_U><OWNER_M>1</OWNER_M><GROUP_U>1</GROUP_U><GROUP_M>1</GROUP_M></PERMISSIONS><CLUSTERS><ID>0</ID></CLUSTERS><TEMPLATE><VN_MAD>` + backendType + `</VN_MAD><MTU>1500</MTU><GUEST_MTU>1500</GUEST_MTU><DESCRIPTION>Preserve me</DESCRIPTION><CUSTOM_TAG>keep</CUSTOM_TAG></TEMPLATE><AR_POOL/></VNET>`
					var escaped bytes.Buffer
					if err := xml.EscapeText(&escaped, []byte(info)); err != nil {
						return nil, err
					}
					value = "<string>" + escaped.String() + "</string>"
				case "one.vn.update":
					updates++
					if call.Params[1].Value.Int != 9 || call.Params[3].Value.Int != 0 {
						return nil, fmt.Errorf("expected replacement update for VNet 9")
					}
					sent := call.Params[2].Value.String
					for _, expected := range []string{`VN_MAD="dummy"`, `DESCRIPTION="Preserve me"`, `CUSTOM_TAG="keep"`, `MTU="1500"`} {
						if !strings.Contains(sent, expected) {
							t.Errorf("update missing %s: %s", expected, sent)
						}
					}
					if strings.Count(sent, "VN_MAD=") != 1 {
						t.Errorf("expected exactly one VN_MAD: %s", sent)
					}
					if reject {
						success = "0"
						value = "<string>Type update rejected</string>"
					} else {
						backendType = "dummy"
						value = "<int>9</int>"
					}
				default:
					return nil, fmt.Errorf("unexpected XML-RPC method %s", call.Method)
				}
				body := `<methodResponse><params><param><value><array><data><value><boolean>` + success + `</boolean></value><value>` + value + `</value><value><int>0</int></value></data></array></value></param></params></methodResponse>`
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			cfg := &Configuration{Controller: goca.NewController(goca.NewClient(goca.NewConfig("test", "test", "http://one.example/RPC2"), &http.Client{Transport: transport}))}
			r := resourceOpennebulaVirtualNetwork()
			old := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "test-vnet", "type": "bridge"})
			old.SetId("9")
			raw := map[string]interface{}{"name": "test-vnet", "type": "dummy"}
			diff, err := r.Diff(context.Background(), old.State(), terraform.NewResourceConfigRaw(raw), cfg)
			if err != nil {
				t.Fatal(err)
			}
			// Only the driver changes; other computed attributes are absent from this fixture.
			for key := range diff.Attributes {
				if key != "type" {
					delete(diff.Attributes, key)
				}
			}
			state, diags := r.Apply(context.Background(), old.State(), diff, cfg)
			if diags.HasError() != reject {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if updates != 1 {
				t.Fatalf("got %d update calls, want 1", updates)
			}
			if !reject {
				if state.ID != "9" || state.Attributes["type"] != "dummy" {
					t.Fatalf("unexpected state after update: ID=%q, type=%q", state.ID, state.Attributes["type"])
				}
			}
		})
	}
}
