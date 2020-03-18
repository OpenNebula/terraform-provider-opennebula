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
			"template": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Description of the template, in OpenNebula's XML or String format",
				Deprecated:  "use other schema sections instead.",
			},
			"cpu":      cpuSchema(),
			"vcpu":     vcpuSchema(),
			"memory":   memorySchema(),
			"context":  contextSchema(),
			"disk":     diskSchema(),
			"graphics": graphicsSchema(),
			"nic":      nicSchema(),
			"os":       osSchema(),
			"vmgroup":  vmGroupSchema(),
			"tags":     tagsSchema(),
		},
	}
}
