package opennebula

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataOpennebulaGroup() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the group",
			},
			"users": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Deprecated: "use user data source for group membership instead.",
			},
			"admins": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Admin user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"quotas": quotasSchema(),
		},
	}
}
