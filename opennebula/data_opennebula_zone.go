package opennebula

import (
	"context"
	"fmt"
	"strconv"

	zoneSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/zone"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaZoneRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Id of the zone",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the zone",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint of the zone",
			},
		},
	}
}

func zoneFilter(d *schema.ResourceData, meta interface{}) (*zoneSc.Zone, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	zones, err := controller.Zones().Info()
	if err != nil {
		return nil, err
	}

	// filter zones with user defined criterias
	id := d.Get("id")
	name, nameOk := d.GetOk("name")

	match := make([]*zoneSc.Zone, 0, 1)
	for i, zone := range zones.Zones {

		if id != -1 && zone.ID != id {
			continue
		}

		if nameOk && zone.Name != name {
			continue
		}

		match = append(match, &zones.Zones[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no zone match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several zones match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	zone, err := zoneFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "zones filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(int64(zone.ID), 10))
	d.Set("name", zone.Name)
	d.Set("endpoint", zone.Template.Endpoint)

	return nil
}
