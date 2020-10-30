package opennebula

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"

	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func quotasSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Define user quota",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"datastore_quotas": {
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
							},
							"size": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Maximum size in MB allowed on the datastore (default: default quota)",
							},
						},
					},
				},
				"network_quotas": {
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
							},
						},
					},
				},
				"image_quotas": {
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
							},
						},
					},
				},
				"vm_quotas": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "VM quotas",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cpu": {
								Type:        schema.TypeFloat,
								Optional:    true,
								Description: "Maximum number of CPU allowed (default: default quota)",
							},
							"memory": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Maximum Memory (MB) allowed (default: default quota)",
							},
							"running_cpu": {
								Type:        schema.TypeFloat,
								Optional:    true,
								Description: "Maximum number of 'running' CPUs allowed (default: default quota)",
							},
							"running_memory": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "'Running' Memory (MB) allowed (default: default quota)",
							},
							"running_vms": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Maximum number of Running VMs allowed (default: default quota)",
							},
							"system_disk_size": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Maximum System Disk size (MB) allowed (default: default quota)",
							},
							"vms": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Maximum number of VMs allowed (default: default quota)",
							},
						},
					},
				},
			},
		},
	}
}

func generateQuotas(d *schema.ResourceData) string {
	quotas := d.Get("quotas").(*schema.Set).List()

	tpl := dyn.NewTemplate()

	quotasMap := quotas[0].(map[string]interface{})
	datastore := quotasMap["datastore_quotas"].([]interface{})
	network := quotasMap["network_quotas"].([]interface{})
	image := quotasMap["image_quotas"].([]interface{})
	vm := quotasMap["vm_quotas"].(*schema.Set).List()

	for i := 0; i < len(datastore); i++ {
		datastoreTpl := tpl.AddVector("DATASTORE")

		datastoreMap := datastore[i].(map[string]interface{})

		datastoreTpl.AddPair("ID", datastoreMap["id"].(int))
		datastoreTpl.AddPair("IMAGES", datastoreMap["images"].(int))
		datastoreTpl.AddPair("SIZE", datastoreMap["size"].(int))
	}

	for i := 0; i < len(network); i++ {
		networkTpl := tpl.AddVector("NETWORK")

		networkMap := network[i].(map[string]interface{})

		networkTpl.AddPair("ID", networkMap["id"].(int))
		networkTpl.AddPair("LEASES", networkMap["leases"].(int))
	}

	for i := 0; i < len(image); i++ {
		imageTpl := tpl.AddVector("IMAGE")

		imageMap := image[i].(map[string]interface{})

		imageTpl.AddPair("ID", imageMap["id"].(int))
		imageTpl.AddPair("RVMS", imageMap["running_vms"].(int))
	}

	if len(vm) > 0 {
		vmMap := vm[0].(map[string]interface{})

		vmTpl := tpl.AddVector("VM")

		vmTpl.AddPair("CPU", float32(vmMap["cpu"].(float64)))
		vmTpl.AddPair("MEMORY", vmMap["memory"].(int))
		vmTpl.AddPair("RUNNING_CPU", float32(vmMap["running_cpu"].(float64)))
		vmTpl.AddPair("RUNNING_MEMORY", vmMap["running_memory"].(int))
		vmTpl.AddPair("RUNNING_VMS", vmMap["running_vms"].(int))
		vmTpl.AddPair("SYSTEM_DISK_SIZE", vmMap["system_disk_size"].(int))
		vmTpl.AddPair("VMS", vmMap["vms"].(int))
	}

	tplStr := tpl.String()

	log.Printf("[INFO] Quotas definition: %s", tplStr)
	return tplStr
}

func flattenQuotasMapFromStructs(d *schema.ResourceData, quotas *shared.QuotasList) error {
	var datastoreQuotas []map[string]interface{}
	var imageQuotas []map[string]interface{}
	var vmQuotas []map[string]interface{}
	var networkQuotas []map[string]interface{}

	// Get datastore quotas
	for _, qds := range quotas.Datastore {
		ds := make(map[string]interface{})
		ds["id"] = qds.ID
		ds["images"] = qds.Images
		ds["size"] = qds.Size
		datastoreQuotas = append(datastoreQuotas, ds)
	}
	// Get network quotas
	for _, qn := range quotas.Network {
		n := make(map[string]interface{})
		n["id"] = qn.ID
		n["leases"] = qn.Leases
		networkQuotas = append(networkQuotas, n)
	}
	// Get VM quotas
	vm := make(map[string]interface{})
	if quotas.VM != nil {
		vm["cpu"] = quotas.VM.CPU
		vm["memory"] = quotas.VM.Memory
		vm["running_cpu"] = quotas.VM.RunningCPU
		vm["running_memory"] = quotas.VM.RunningMemory
		vm["vms"] = quotas.VM.VMs
		vm["running_vms"] = quotas.VM.RunningVMs
		vm["system_disk_size"] = quotas.VM.SystemDiskSize
		vmQuotas = append(vmQuotas, vm)
	}
	// Get Image quotas
	for _, qimg := range quotas.Image {
		img := make(map[string]interface{})
		img["id"] = qimg.ID
		img["running_vms"] = qimg.RVMs
		imageQuotas = append(imageQuotas, img)
	}

	return d.Set("quotas", []interface{}{
		map[string]interface{}{
			"datastore_quotas": datastoreQuotas,
			"image_quotas":     imageQuotas,
			"vm_quotas":        vmQuotas,
			"network_quotas":   networkQuotas,
		},
	})
}
