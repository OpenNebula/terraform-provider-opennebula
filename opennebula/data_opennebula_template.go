package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataOpennebulaTemplate() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaTemplateRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Template",
			},
		},
	}
}
