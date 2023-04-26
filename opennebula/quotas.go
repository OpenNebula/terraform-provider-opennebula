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
)

func quotasMapSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		},
	}
}

func generateQuotas(d *schema.ResourceData, forceDefault bool) (string, error) {

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

func flattenQuotasMapFromStructs(d *schema.ResourceData, quotas *shared.QuotasList) error {
	var datastoreQuotas []map[string]interface{}
	var imageQuotas []map[string]interface{}
	var vmQuotas []map[string]interface{}
	var networkQuotas []map[string]interface{}
	var q uint8
	q = 0

	// Get datastore quotas
	for _, qds := range quotas.Datastore {
		ds := make(map[string]interface{})
		ds["id"] = qds.ID
		ds["images"] = qds.Images
		ds["size"] = qds.Size
		if len(ds) > 0 {
			datastoreQuotas = append(datastoreQuotas, ds)
		}
		q = q | DSFlag
	}

	// Get network quotas
	for _, qn := range quotas.Network {
		n := make(map[string]interface{})
		n["id"] = qn.ID
		n["leases"] = qn.Leases
		if len(n) > 0 {
			networkQuotas = append(networkQuotas, n)
		}
		q = q | NetFlag
	}
	// Get VM quotas
	if quotas.VM != nil {
		vm := make(map[string]interface{})

		vm["cpu"] = quotas.VM.CPU
		vm["memory"] = quotas.VM.Memory
		vm["running_cpu"] = quotas.VM.RunningCPU
		vm["running_memory"] = quotas.VM.RunningMemory
		vm["vms"] = quotas.VM.VMs
		vm["running_vms"] = quotas.VM.RunningVMs
		vm["system_disk_size"] = quotas.VM.SystemDiskSize

		if len(vm) > 0 {
			vmQuotas = append(vmQuotas, vm)
		}
		q = q | VMFlag
	}
	// Get Image quotas
	for _, qimg := range quotas.Image {
		img := make(map[string]interface{})
		img["id"] = qimg.ID
		img["running_vms"] = qimg.RVMs
		if len(img) > 0 {
			imageQuotas = append(imageQuotas, img)
		}
		q = q | ImgFlag
	}

	for q > 0 {
		switch {
		case q&DSFlag > 0:
			d.Set("datastore", datastoreQuotas)
			q = q ^ DSFlag
		case q&NetFlag > 0:
			d.Set("network", networkQuotas)
			q = q ^ NetFlag
		case q&VMFlag > 0:
			d.Set("vm", vmQuotas)
			q = q ^ VMFlag
		case q&ImgFlag > 0:
			d.Set("image", imageQuotas)
			q = q ^ ImgFlag
		}
	}

	return nil
}
