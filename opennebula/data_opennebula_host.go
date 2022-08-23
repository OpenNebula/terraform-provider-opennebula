package opennebula

import (
	"context"
	"fmt"
	"strconv"

	hostSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/host"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaHost() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaHostRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the host",
			},
			"tags": tagsSchema(),
		},
	}
}

func hostFilter(d *schema.ResourceData, meta interface{}) (*hostSc.Host, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	hosts, err := controller.Hosts().Info()
	if err != nil {
		return nil, err
	}

	// filter hosts with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*hostSc.Host, 0, 1)
	for i, host := range hosts.Hosts {

		if nameOk && host.Name != name {
			continue
		}

		if tagsOk && !matchTags(host.Template.Template, tags) {
			continue
		}

		match = append(match, &hosts.Hosts[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no host match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several hosts match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	host, err := hostFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "hosts filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(host.Template.Template)

	d.SetId(strconv.FormatInt(int64(host.ID), 10))
	d.Set("name", host.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("Host (ID: %d): %s", host.ID, err),
			})
			return diags
		}
	}

	return nil
}
