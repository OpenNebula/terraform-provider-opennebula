package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataOpennebulaVMGroup() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaVMGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Virtual Machine Group",
			},
			"tags": tagsSchema(),
		},
	}
}
