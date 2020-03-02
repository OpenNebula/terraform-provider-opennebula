package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataOpennebulaImage() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaImageRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Image",
			},
			"tags": tagsSchema(),
		},
	}
}
