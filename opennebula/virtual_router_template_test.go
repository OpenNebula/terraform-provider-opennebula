package opennebula

import (
	"encoding/xml"
	"testing"

	vr "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualrouter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFlattenVirtualRouterMultipleNICs(t *testing.T) {
	for _, tc := range []struct {
		name           string
		nics           string
		wantCount      int
		wantManagement string
		wantError      bool
	}{
		{name: "target second", nics: `<NIC><NETWORK_ID>49</NETWORK_ID></NIC><NIC><NIC_ID>1</NIC_ID><NETWORK_ID>10</NETWORK_ID><VROUTER_MANAGEMENT>YES</VROUTER_MANAGEMENT></NIC>`, wantCount: 1, wantManagement: "YES"},
		{name: "actual flag differs", nics: `<NIC><NETWORK_ID>49</NETWORK_ID></NIC><NIC><NETWORK_ID>10</NETWORK_ID><VROUTER_MANAGEMENT>NO</VROUTER_MANAGEMENT></NIC>`, wantCount: 1, wantManagement: "NO"},
		{name: "target absent", nics: `<NIC><NETWORK_ID>49</NETWORK_ID></NIC><NIC><NETWORK_ID>50</NETWORK_ID></NIC>`},
		{name: "ambiguous network", nics: `<NIC><NETWORK_ID>10</NETWORK_ID></NIC><NIC><NETWORK_ID>10</NETWORK_ID></NIC>`, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var template vr.Template
			if err := xml.Unmarshal([]byte(`<TEMPLATE>`+tc.nics+`</TEMPLATE>`), &template); err != nil {
				t.Fatal(err)
			}
			d := schema.TestResourceDataRaw(t, resourceOpennebulaVirtualRouter().Schema, map[string]interface{}{
				"name": "router", "instance_template_id": 0,
				"template_section": []interface{}{map[string]interface{}{"name": "NIC", "elements": map[string]interface{}{"NETWORK_ID": "10", "VROUTER_MANAGEMENT": "YES"}}},
			})
			diags := flattenVirtualRouterTemplate(d, &Configuration{}, &template)
			if diags.HasError() != tc.wantError {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if tc.wantError {
				return
			}
			sections := d.Get("template_section").(*schema.Set).List()
			if len(sections) != tc.wantCount {
				t.Fatalf("got %d sections, want %d", len(sections), tc.wantCount)
			}
			if tc.wantCount > 0 {
				elements := sections[0].(map[string]interface{})["elements"].(map[string]interface{})
				if elements["NETWORK_ID"] != "10" || elements["VROUTER_MANAGEMENT"] != tc.wantManagement {
					t.Fatalf("unexpected NIC: %#v", elements)
				}
			}
		})
	}
}

func TestFlattenVirtualRouterDescriptionAndNIC(t *testing.T) {
	const description = "FRR vRouter pair"
	var template vr.Template
	if err := xml.Unmarshal([]byte(`<TEMPLATE><DESCRIPTION><![CDATA[`+description+`]]></DESCRIPTION><NIC><NETWORK_ID><![CDATA[49]]></NETWORK_ID><VROUTER_MANAGEMENT><![CDATA[YES]]></VROUTER_MANAGEMENT></NIC></TEMPLATE>`), &template); err != nil {
		t.Fatal(err)
	}
	d := schema.TestResourceDataRaw(t, resourceOpennebulaVirtualRouter().Schema, map[string]interface{}{
		"name": "router", "instance_template_id": 0,
		"template_section": []interface{}{map[string]interface{}{"name": "NIC", "elements": map[string]interface{}{"NETWORK_ID": "49", "VROUTER_MANAGEMENT": "YES"}}},
	})
	if diags := flattenVirtualRouterTemplate(d, &Configuration{}, &template); diags.HasError() {
		t.Fatal(diags)
	}
	if got := d.Get("description"); got != description {
		t.Fatalf("description = %q, want %q", got, description)
	}
	sections := d.Get("template_section").(*schema.Set).List()
	if len(sections) != 1 {
		t.Fatalf("got %d template sections, want 1", len(sections))
	}
	elements := sections[0].(map[string]interface{})["elements"].(map[string]interface{})
	if elements["NETWORK_ID"] != "49" || elements["VROUTER_MANAGEMENT"] != "YES" {
		t.Fatalf("unexpected NIC: %#v", elements)
	}
}
