package opennebula

import (
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
				Description: "Description of the vnet, in OpenNebula's XML or String format",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "MTU of the vnet (defaut: 1500)",
			},
		},
	}
}
