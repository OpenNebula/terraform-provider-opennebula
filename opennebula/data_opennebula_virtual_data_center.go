package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataOpennebulaVirtualDataCenter() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaVirtualDataCenterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC",
			},
		},
	}
}
