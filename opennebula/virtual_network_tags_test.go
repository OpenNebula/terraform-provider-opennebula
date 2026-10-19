package opennebula

import (
	"encoding/xml"
	"testing"

	vn "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFlattenVnetDescriptionTag(t *testing.T) {
	const description = "Bridged FaaS IPv6 guest network on brfaas"
	var template vn.Template
	if err := xml.Unmarshal([]byte(`<TEMPLATE><DESCRIPTION><![CDATA[`+description+`]]></DESCRIPTION><GUEST_MTU><![CDATA[1500]]></GUEST_MTU><MTU><![CDATA[1500]]></MTU></TEMPLATE>`), &template); err != nil {
		t.Fatal(err)
	}
	d := schema.TestResourceDataRaw(t, resourceOpennebulaVirtualNetwork().Schema, map[string]interface{}{
		"name": "faas-vnet", "tags": map[string]interface{}{"DESCRIPTION": ""},
	})
	flattenVnetTemplate(d, &Configuration{}, false, &template)
	for _, field := range []string{"tags", "tags_all"} {
		if got := d.Get(field).(map[string]interface{})["DESCRIPTION"]; got != description {
			t.Errorf("%s DESCRIPTION = %q, want %q", field, got, description)
		}
	}
}
