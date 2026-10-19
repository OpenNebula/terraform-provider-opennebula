package opennebula

import (
	"encoding/xml"
	"testing"

	vr "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualrouter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
