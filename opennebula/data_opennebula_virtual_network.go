package opennebula

import (
	"fmt"
	"strconv"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	vnetSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataOpennebulaVirtualNetwork() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaVirtualNetworkRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Virtual Network",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "use 'tags' for selection instead",
				Description: "Description of the vnet, in OpenNebula's XML or String format",
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

	controller := meta.(*goca.Controller)

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
		return nil, fmt.Errorf("no virtual network match the tags")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several virtual networks match the tags")
	}

	return match[0], nil
}

func datasourceOpennebulaVirtualNetworkRead(d *schema.ResourceData, meta interface{}) error {

	vnet, err := vnetFilter(d, meta)
	if err != nil {
		return err
	}

	tplPairs := pairsToMap(vnet.Template.Template)

	d.SetId(strconv.FormatInt(int64(vnet.ID), 10))
	d.Set("name", vnet.Name)

	mtu, err := vnet.Template.GetI("MTU")
	if err != nil {
		return err
	}
	d.Set("mtu", mtu)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			return err
		}
	}

	return nil
}
