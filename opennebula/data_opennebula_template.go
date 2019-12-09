package opennebula

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
			"template": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "template content",
			},
		},
	}
}
