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
			},
			"admins": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Admin user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"quotas": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Define group quota",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datastore_quotas": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Datastore quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Datastore ID",
									},
									"images": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of Images allowed (default: default quota)",
									},
									"size": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum size in MB allowed on the datastore (default: default quota)",
									},
								},
							},
						},
						"network_quotas": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Network quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Network ID",
									},
									"leases": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of Leases allowed for this network (default: default quota)",
									},
								},
							},
						},
						"image_quotas": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Image quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Image ID",
									},
									"running_vms": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of Running VMs allowed for this image (default: default quota)",
									},
								},
							},
						},
						"vm_quotas": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "VM quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of CPU allowed (default: default quota)",
									},
									"memory": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum Memory (MB) allowed (default: default quota)",
									},
									"running_cpu": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of 'running' CPUs allowed (default: default quota)",
									},
									"running_memory": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "'Running' Memory (MB) allowed (default: default quota)",
									},
									"running_vms": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of Running VMs allowed (default: default quota)",
									},
									"system_disk_size": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum System Disk size (MB) allowed (default: default quota)",
									},
									"vms": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Maximum number of VMs allowed (default: default quota)",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
