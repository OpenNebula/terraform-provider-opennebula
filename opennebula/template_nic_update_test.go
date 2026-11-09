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

type templateNICUpdateTransport func(*http.Request) (*http.Response, error)

func (f templateNICUpdateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTemplateNICOnlyUpdateIsSent(t *testing.T) {
	for _, tc := range []struct {
		field    string
		value    map[string]interface{}
		expected []string
	}{
		{field: "nic", value: map[string]interface{}{"network_mode_auto": true, "security_groups": []interface{}{100}}, expected: []string{`NETWORK_MODE="auto"`, `SECURITY_GROUPS="100"`}},
		{field: "nic_alias", value: map[string]interface{}{"parent": "eth0", "network_id": 10}, expected: []string{`PARENT="eth0"`, `NETWORK_ID="10"`}},
	} {
		t.Run(tc.field, func(t *testing.T) {
			updates := 0
			sent := ""
			transport := templateNICUpdateTransport(func(req *http.Request) (*http.Response, error) {
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
				case "one.template.info":
					var escaped bytes.Buffer
					if err := xml.EscapeText(&escaped, []byte(`<VMTEMPLATE><ID>1</ID><NAME>sles-15-sp7</NAME><TEMPLATE><CPU>1</CPU><MEMORY>512</MEMORY></TEMPLATE></VMTEMPLATE>`)); err != nil {
						return nil, err
					}
					value = "<string>" + escaped.String() + "</string>"
				case "one.template.update":
					updates++
					sent = call.Params[2].Value.String
					value = "<int>1</int>"
				default:
					return nil, fmt.Errorf("unexpected XML-RPC method %s", call.Method)
				}
				body := `<methodResponse><params><param><value><array><data><value><boolean>1</boolean></value><value>` + value + `</value><value><int>0</int></value></data></array></value></param></params></methodResponse>`
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			cfg := &Configuration{Controller: goca.NewController(goca.NewClient(goca.NewConfig("test", "test", "http://one.example/RPC2"), &http.Client{Transport: transport}))}
			r := resourceOpennebulaTemplate()
			// Exercise the real update path without the subsequent unrelated refresh.
			r.UpdateContext = resourceOpennebulaTemplateUpdateCustom
			old := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"name": "sles-15-sp7"})
			old.SetId("1")
			raw := map[string]interface{}{"name": "sles-15-sp7", tc.field: []interface{}{tc.value}}
			diff, err := r.Diff(context.Background(), old.State(), terraform.NewResourceConfigRaw(raw), cfg)
			if err != nil {
				t.Fatal(err)
			}
			// Isolate the NIC change from computed fields absent in the minimal fixture.
			for key := range diff.Attributes {
				if !strings.HasPrefix(key, tc.field+".") {
					delete(diff.Attributes, key)
				}
			}
			if _, diags := r.Apply(context.Background(), old.State(), diff, cfg); diags.HasError() {
				t.Fatal(diags)
			}
			if updates != 1 {
				t.Fatalf("got %d template update calls, want 1", updates)
			}
			for _, expected := range tc.expected {
				if !strings.Contains(sent, expected) {
					t.Errorf("update template missing %s: %s", expected, sent)
				}
			}
		})
	}
}
