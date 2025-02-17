package opennebula

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

const (
	DefaultQuotaValue = -1

	DSFlag  uint8 = 1
	NetFlag       = 2
	VMFlag        = 4
	ImgFlag       = 8

	DatastoreQuota = "datastore"
	NetworkQuota   = "network"
	ImageQuota     = "image"
	VMQuota        = "vm"
)

var validQuotaTypes = []string{DatastoreQuota, NetworkQuota, ImageQuota, VMQuota}

func quotasMapSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": {
			Type:     schema.TypeString,
			Optional: false,
			Computed: true,
		},
		"datastore": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Datastore quotas",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Datastore ID",
					},
					"images": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum number of Images allowed (default: default quota)",
						Default:     -1,
					},
					"size": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum size in MB allowed on the datastore (default: default quota)",
						Default:     DefaultQuotaValue,
					},
				},
			},
			ConflictsWith: []string{"network", "image", "vm"},
		},
		"network": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Network quotas",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Network ID",
					},
					"leases": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum number of Leases allowed for this network (default: default quota)",
						Default:     DefaultQuotaValue,
					},
				},
			},
			ConflictsWith: []string{"datastore", "image", "vm"},
		},
		"image": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Image quotas",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Image ID",
					},
					"running_vms": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum number of Running VMs allowed for this image (default: default quota)",
						Default:     DefaultQuotaValue,
					},
				},
			},
			ConflictsWith: []string{"datastore", "network", "vm"},
		},
		"vm": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "VM quotas",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cpu": {
						Type:        schema.TypeFloat,
						Optional:    true,
						Description: "Maximum number of CPU allowed (default: default quota)",
						Default:     DefaultQuotaValue,
					},
					"memory": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum Memory (MB) allowed (default: default quota)",
						Default:     DefaultQuotaValue,
					},
					"running_cpu": {
						Type:        schema.TypeFloat,
						Optional:    true,
						Description: "Maximum number of 'running' CPUs allowed (default: default quota)",
						Default:     DefaultQuotaValue,
					},
					"running_memory": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "'Running' Memory (MB) allowed (default: default quota)",
						Default:     DefaultQuotaValue,
					},
					"running_vms": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum number of Running VMs allowed (default: default quota)",
						Default:     DefaultQuotaValue,
					},
					"system_disk_size": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum System Disk size (MB) allowed (default: default quota)",
						Default:     DefaultQuotaValue,
					},
					"vms": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum number of VMs allowed (default: default quota)",
						Default:     DefaultQuotaValue,
					},
				},
			},
			ConflictsWith: []string{"datastore", "network", "image"},
		},
	}
}

func generateQuotas(d *schema.ResourceData, forceDefault bool) (string, error) {

	// TODO: check if type match section

	tpl := dyn.NewTemplate()

	datastore := d.Get("datastore").([]interface{})
	network := d.Get("network").([]interface{})
	image := d.Get("image").([]interface{})
	vm := d.Get("vm").([]interface{})

	for i := 0; i < len(datastore); i++ {
		datastoreTpl := tpl.AddVector("DATASTORE")

		datastoreMap := datastore[i].(map[string]interface{})

		datastoreTpl.AddPair("ID", datastoreMap["id"].(int))
		if forceDefault {
			datastoreTpl.AddPair("IMAGES", DefaultQuotaValue)
			datastoreTpl.AddPair("SIZE", DefaultQuotaValue)
		} else {
			datastoreTpl.AddPair("IMAGES", datastoreMap["images"].(int))
			datastoreTpl.AddPair("SIZE", datastoreMap["size"].(int))
		}
	}

	for i := 0; i < len(network); i++ {
		networkTpl := tpl.AddVector("NETWORK")

		networkMap := network[i].(map[string]interface{})

		networkTpl.AddPair("ID", networkMap["id"].(int))

		if forceDefault {
			networkTpl.AddPair("LEASES", DefaultQuotaValue)
		} else {
			networkTpl.AddPair("LEASES", networkMap["leases"].(int))
		}
	}

	for i := 0; i < len(image); i++ {
		imageTpl := tpl.AddVector("IMAGE")

		imageMap := image[i].(map[string]interface{})

		imageTpl.AddPair("ID", imageMap["id"].(int))

		if forceDefault {
			imageTpl.AddPair("RVMS", DefaultQuotaValue)
		} else {
			imageTpl.AddPair("RVMS", imageMap["running_vms"].(int))
		}
	}

	if len(vm) > 0 {
		vmMap := vm[0].(map[string]interface{})

		vmTpl := tpl.AddVector("VM")

		if forceDefault {
			vmTpl.AddPair("CPU", DefaultQuotaValue)
			vmTpl.AddPair("MEMORY", DefaultQuotaValue)
			vmTpl.AddPair("RUNNING_CPU", DefaultQuotaValue)
			vmTpl.AddPair("RUNNING_MEMORY", DefaultQuotaValue)
			vmTpl.AddPair("RUNNING_VMS", DefaultQuotaValue)
			vmTpl.AddPair("SYSTEM_DISK_SIZE", DefaultQuotaValue)
			vmTpl.AddPair("VMS", DefaultQuotaValue)
		} else {
			vmTpl.AddPair("CPU", float32(vmMap["cpu"].(float64)))
			vmTpl.AddPair("MEMORY", vmMap["memory"].(int))
			vmTpl.AddPair("RUNNING_CPU", float32(vmMap["running_cpu"].(float64)))
			vmTpl.AddPair("RUNNING_MEMORY", vmMap["running_memory"].(int))
			vmTpl.AddPair("RUNNING_VMS", vmMap["running_vms"].(int))
			vmTpl.AddPair("SYSTEM_DISK_SIZE", vmMap["system_disk_size"].(int))
			vmTpl.AddPair("VMS", vmMap["vms"].(int))
		}
	}

	tplStr := tpl.String()

	log.Printf("[INFO] Quotas definition: %s", tplStr)
	return tplStr, nil
}

func flattenDatastoreQuota(d *schema.ResourceData, quotas []shared.DatastoreQuota) error {
	var datastoreQuotas []map[string]interface{}

	for _, qds := range quotas {
		ds := make(map[string]interface{})
		ds["id"] = qds.ID
		ds["images"] = qds.Images
		ds["size"] = qds.Size
		if len(ds) > 0 {
			datastoreQuotas = append(datastoreQuotas, ds)
		}
	}
	return d.Set("datastore", datastoreQuotas)
}

func flattenNetworkQuota(d *schema.ResourceData, quotas []shared.NetworkQuota) error {
	var networkQuotas []map[string]interface{}

	// Get network quotas
	for _, qn := range quotas {
		n := make(map[string]interface{})
		n["id"] = qn.ID
		n["leases"] = qn.Leases
		if len(n) > 0 {
			networkQuotas = append(networkQuotas, n)
		}
	}
	return d.Set("network", networkQuotas)
}

func flattenVMQuota(d *schema.ResourceData, quotas []shared.VMQuota) error {
	var vmQuotas []map[string]interface{}

	// Get VM quotas
	for _, qvm := range quotas {
		vm := make(map[string]interface{})
		vm["cpu"] = qvm.CPU
		vm["memory"] = qvm.Memory
		vm["running_cpu"] = qvm.RunningCPU
		vm["running_memory"] = qvm.RunningMemory
		vm["vms"] = qvm.VMs
		vm["running_vms"] = qvm.RunningVMs
		vm["system_disk_size"] = qvm.SystemDiskSize
		if len(vm) > 0 {
			vmQuotas = append(vmQuotas, vm)
		}
	}
	return d.Set("vm", vmQuotas)
}

func flattenImageQuota(d *schema.ResourceData, quotas []shared.ImageQuota) error {
	var imageQuotas []map[string]interface{}

	// Get Image quotas
	for _, qimg := range quotas {
		img := make(map[string]interface{})
		img["id"] = qimg.ID
		img["running_vms"] = qimg.RVMs
		if len(img) > 0 {
			imageQuotas = append(imageQuotas, img)
		}
	}
	return d.Set("image", imageQuotas)
}

func flattenQuotasMapFromStructs(d *schema.ResourceData, quotas *shared.QuotasList) error {

	// quotas resources defines only one kind of resource at a time
	log.Printf("[INFO] === type: %s\n", d.Get("type").(string))

	// what if type is not defined ?
	quotasType := d.Get("type").(string)
	if len(quotasType) == 0 {
		if len(d.Get("datastore").([]interface{})) > 0 {
			quotasType = DatastoreQuota
		} else if len(d.Get("network").([]interface{})) > 0 {
			quotasType = NetworkQuota
		} else if len(d.Get("image").([]interface{})) > 0 {
			quotasType = ImageQuota
		} else if len(d.Get("vm").([]interface{})) > 0 {
			quotasType = VMQuota
		}
	}

	switch quotasType {
	case DatastoreQuota:
		flattenDatastoreQuota(d, quotas.Datastore)
	case NetworkQuota:
		flattenNetworkQuota(d, quotas.Network)
	case ImageQuota:
		flattenImageQuota(d, quotas.Image)
	case VMQuota:
		flattenVMQuota(d, quotas.VM)
	}

	return nil
}
