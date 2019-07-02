package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataOpennebulaGroup() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Group",
			},
		},
	}
}
