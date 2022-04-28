package opennebula

import (
	"context"
	"fmt"
	"strconv"

	vmGroupSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vmgroup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaVMGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaVMGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Virtual Machine Group",
			},
			"tags": tagsSchema(),
		},
	}
}

func vmGroupFilter(d *schema.ResourceData, meta interface{}) (*vmGroupSc.VMGroup, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	// search for any vmgroups this user can see
	vmGroups, err := controller.VMGroups().Info(-2, -1, -1)
	if err != nil {
		return nil, err
	}

	// filter vm groups with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*vmGroupSc.VMGroup, 0, 1)
	for i, vmGroup := range vmGroups.VMGroups {

		if nameOk && vmGroup.Name != name {
			continue
		}

		if tagsOk && !matchTags(vmGroup.Template, tags) {
			continue
		}

		match = append(match, &vmGroups.VMGroups[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no vm group match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several vm group match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaVMGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vmGroup, err := vmGroupFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "VM groups filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(vmGroup.Template)

	d.SetId(strconv.FormatInt(int64(vmGroup.ID), 10))
	d.Set("name", vmGroup.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("VM group (ID: %d): %s", vmGroup.ID, err),
			})
			return diags
		}
	}

	return nil
}
