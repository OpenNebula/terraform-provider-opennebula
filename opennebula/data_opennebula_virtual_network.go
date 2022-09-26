package opennebula

import (
	"context"
	"fmt"
	"strconv"

	vnetSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaVirtualNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaVirtualNetworkRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Virtual Network",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "MTU of the vnet (defaut: 1500)",
			},
			"tags": tagsSchema(),
		},
	}
}

func vnetFilter(d *schema.ResourceData, meta interface{}) (*vnetSc.VirtualNetwork, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	vnets, err := controller.VirtualNetworks().Info()
	if err != nil {
		return nil, err
	}

	// filter vnets with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*vnetSc.VirtualNetwork, 0, 1)
	for i, vnet := range vnets.VirtualNetworks {

		if nameOk && vnet.Name != name {
			continue
		}

		if tagsOk && !matchTags(vnet.Template.Template, tags) {
			continue
		}

		match = append(match, &vnets.VirtualNetworks[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no virtual network match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several virtual networks match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaVirtualNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vnet, err := vnetFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual networks filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(vnet.Template.Template)

	d.SetId(strconv.FormatInt(int64(vnet.ID), 10))
	d.Set("name", vnet.Name)

	mtu, err := vnet.Template.GetI("MTU")
	if err == nil {
		d.Set("mtu", mtu)
	}

	err = d.Set("tags", tplPairs)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "setting attribute failed",
			Detail:   fmt.Sprintf("Virtual network (ID: %d): %s", vnet.ID, err),
		})
		return diags
	}

	return nil
}
