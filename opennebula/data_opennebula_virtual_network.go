package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
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
		},
	}
}
