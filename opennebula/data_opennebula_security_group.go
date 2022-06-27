package opennebula

import (
	"context"
	"fmt"
	"strconv"

	secgroup "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/securitygroup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaSecurityGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaSecurityGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Security Group",
			},
			"tags": tagsSchema(),
		},
	}
}

func securityGroupFilter(d *schema.ResourceData, meta interface{}) (*secgroup.SecurityGroup, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	securityGroups, err := controller.SecurityGroups().Info()
	if err != nil {
		return nil, err
	}

	// filter security groups with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*secgroup.SecurityGroup, 0, 1)
	for i, securityGroup := range securityGroups.SecurityGroups {

		if nameOk && securityGroup.Name != name {
			continue
		}

		if tagsOk && !matchTags(securityGroup.Template.Template, tags) {
			continue
		}

		match = append(match, &securityGroups.SecurityGroups[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no security group match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several security group match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	securityGroup, err := securityGroupFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "security groups filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(securityGroup.Template.Template)

	d.SetId(strconv.FormatInt(int64(securityGroup.ID), 10))
	d.Set("name", securityGroup.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("Security group (ID: %d): %s", securityGroup.ID, err),
			})
			return diags
		}
	}

	return nil
}
