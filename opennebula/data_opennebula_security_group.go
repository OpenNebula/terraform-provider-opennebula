package opennebula

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataOpennebulaSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaSecurityGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Security Group",
			},
		},
	}
}
