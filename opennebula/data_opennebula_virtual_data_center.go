package opennebula

import (
	"context"
	"fmt"
	"strconv"

	vdcSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vdc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaVirtualDataCenter() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaVirtualDataCenterRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Id of the datacenter",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the VDC",
			},
			"tags": tagsSchema(),
		},
	}
}

func vdcFilter(d *schema.ResourceData, meta interface{}) (*vdcSc.VDC, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	vdcs, err := controller.VDCs().Info()
	if err != nil {
		return nil, err
	}

	// filter vdcs with user defined criterias
	id := d.Get("id")
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*vdcSc.VDC, 0, 1)
	for i, vdc := range vdcs.VDCs {

		if id != -1 && vdc.ID != id {
			continue
		}

		if nameOk && vdc.Name != name {
			continue
		}

		if tagsOk && !matchTags(vdc.Template.Template, tags) {
			continue
		}

		match = append(match, &vdcs.VDCs[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no virtual data center match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several virtual data centers match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaVirtualDataCenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vdc, err := vdcFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual data centers filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(vdc.Template.Template)

	d.SetId(strconv.FormatInt(int64(vdc.ID), 10))
	d.Set("name", vdc.Name)

	err = d.Set("tags", tplPairs)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "setting attribute failed",
			Detail:   fmt.Sprintf("Virtual data center (ID: %d): %s", vdc.ID, err),
		})
		return diags
	}

	return nil
}
