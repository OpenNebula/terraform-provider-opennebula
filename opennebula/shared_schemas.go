package opennebula

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

func nicSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		Description: "Definition of network adapter(s) assigned to the Virtual Machine",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ip": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"mac": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"model": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"network_id": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"network": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"physical_device": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"security_groups": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeInt,
					},
				},
				"nic_id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}
}

func diskSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		Description: "Definition of disks assigned to the Virtual Machine",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"image_id": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"size": {
					Type:     schema.TypeInt,
					Computed: true,
					Optional: true,
				},
				"target": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"driver": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
			},
		},
	}
}

func contextSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Optional:    true,
		Description: "Context variables",
	}
}

func graphicsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Description: "Definition of graphics adapter assigned to the Virtual Machine",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"listen": {
					Type:     schema.TypeString,
					Required: true,
				},
				"port": {
					Type:     schema.TypeString,
					Computed: true,
					Optional: true,
				},
				"type": {
					Type:     schema.TypeString,
					Required: true,
				},
				"keymap": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "en-us",
				},
			},
		},
	}
}

func osSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Description: "Definition of OS boot and type for the Virtual Machine",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arch": {
					Type:     schema.TypeString,
					Required: true,
				},
				"boot": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func cpuSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeFloat,
		Optional:    true,
		Computed:    true,
		Description: "Amount of CPU quota assigned to the virtual machine",
	}
}

func vcpuSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		Description: "Number of virtual CPUs assigned to the virtual machine",
	}
}

func memorySchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		Description: "Amount of memory (RAM) in MB assigned to the virtual machine",
	}
}

func vmGroupSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		ForceNew:    true,
		MaxItems:    1,
		Description: "Virtual Machine Group to associate with during VM creation only. If it changes, a New VM is created",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"vmgroup_id": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"role": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func tagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Optional:    true,
		Computed:    true,
		Description: "Add custom tags to the resource",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func generateVMTemplate(d *schema.ResourceData, tpl *vm.Template) {

	//Generate NIC definition
	nics := d.Get("nic").([]interface{})
	log.Printf("Number of NICs: %d", len(nics))

	for i := 0; i < len(nics); i++ {
		nicconfig := nics[i].(map[string]interface{})
		nic := tpl.AddNIC()

		for k, v := range nicconfig {

			if k == "network_id" {
				nic.Add(shared.NetworkID, strconv.Itoa(v.(int)))
				continue
			}

			if isEmptyValue(reflect.ValueOf(v)) {
				continue
			}

			switch k {
			case "ip":
				nic.Add(shared.IP, v.(string))
			case "mac":
				nic.Add(shared.MAC, v.(string))
			case "model":
				nic.Add(shared.Model, v.(string))
			case "physical_device":
				nic.Add("PHYDEV", v.(string))
			case "security_groups":
				nicsecgroups := ArrayToString(v.([]interface{}), ",")
				nic.Add(shared.SecurityGroups, nicsecgroups)
			}
		}

	}

	//Generate DISK definition
	disks := d.Get("disk").([]interface{})
	log.Printf("Number of disks: %d", len(disks))

	for i := 0; i < len(disks); i++ {

		diskconfig := disks[i].(map[string]interface{})
		disk := tpl.AddDisk()

		for k, v := range diskconfig {

			if isEmptyValue(reflect.ValueOf(v)) {
				continue
			}

			switch k {
			case "target":
				disk.Add(shared.TargetDisk, v.(string))
			case "driver":
				disk.Add(shared.Driver, v.(string))
			case "size":
				disk.Add(shared.Size, strconv.Itoa(v.(int)))
			case "image_id":
				disk.Add(shared.ImageID, strconv.Itoa(v.(int)))
			}
		}
	}

	//Generate GRAPHICS definition
	graphics := d.Get("graphics").([]interface{})
	for i := 0; i < len(graphics); i++ {
		graphicsconfig := graphics[i].(map[string]interface{})

		for k, v := range graphicsconfig {

			if isEmptyValue(reflect.ValueOf(v)) {
				continue
			}

			switch k {
			case "listen":
				tpl.AddIOGraphic(vmk.Listen, v.(string))
			case "type":
				tpl.AddIOGraphic(vmk.GraphicType, v.(string))
			case "port":
				tpl.AddIOGraphic(vmk.Port, v.(string))
			case "keymap":
				tpl.AddIOGraphic(vmk.Keymap, v.(string))
			}

		}
	}

	//Generate OS definition
	os := d.Get("os").([]interface{})
	for i := 0; i < len(os); i++ {
		osconfig := os[i].(map[string]interface{})
		tpl.AddOS(vmk.Arch, osconfig["arch"].(string))
		tpl.AddOS(vmk.Boot, osconfig["boot"].(string))
	}

	//Generate VM Group definition
	vmgroup := d.Get("vmgroup").([]interface{})
	for i := 0; i < len(vmgroup); i++ {
		vmgconfig := vmgroup[i].(map[string]interface{})
		vmgroupTpl := tpl.AddVector("VMGROUP")
		vmgroupTpl.AddPair("VMGROUP_ID", vmgconfig["vmgroup_id"].(int))
		vmgroupTpl.AddPair("ROLE", vmgconfig["role"].(string))
	}

	vmcpu, ok := d.GetOk("cpu")
	if ok {
		tpl.CPU(vmcpu.(float64))
	}
	vmmemory, ok := d.GetOk("memory")
	if ok {
		tpl.Memory(vmmemory.(int))
	}
	vmvcpu, ok := d.GetOk("vcpu")
	if ok {
		tpl.VCPU(vmvcpu.(int))
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
	}

}

func flattenTemplate(d *schema.ResourceData, vmTemplate *vm.Template, tplTags bool) error {

	var err error

	// VM Group
	vmgMap := make([]map[string]interface{}, 0, 1)
	vmgIdStr, _ := vmTemplate.GetStrFromVec("VMGROUP", "VMGROUP_ID")
	vmgid, _ := strconv.ParseInt(vmgIdStr, 10, 32)
	vmgRole, _ := vmTemplate.GetStrFromVec("VMGROUP", "ROLE")

	// OS
	osMap := make([]map[string]interface{}, 0, 1)
	arch, _ := vmTemplate.GetOS(vmk.Arch)
	boot, _ := vmTemplate.GetOS(vmk.Boot)

	// Graphics
	graphMap := make([]map[string]interface{}, 0, 1)
	listen, _ := vmTemplate.GetIOGraphic(vmk.Listen)
	port, _ := vmTemplate.GetIOGraphic(vmk.Port)
	t, _ := vmTemplate.GetIOGraphic(vmk.GraphicType)
	keymap, _ := vmTemplate.GetIOGraphic(vmk.Keymap)

	// Disks
	diskList := make([]interface{}, 0, 1)

	// Nics
	nicList := make([]interface{}, 0, 1)

	// Context
	context := make(map[string]interface{})
	vmcontext, _ := vmTemplate.GetVector(vmk.ContextVec)

	// Set VM Group to resource
	if vmgIdStr != "" {
		vmgMap = append(vmgMap, map[string]interface{}{
			"vmgroup_id": vmgid,
			"role":       vmgRole,
		})
		err = d.Set("vmgroup", vmgMap)
		if err != nil {
			return err
		}
	}

	// Set OS to resource
	if arch != "" {
		osMap = append(osMap, map[string]interface{}{
			"arch": arch,
			"boot": boot,
		})
		err = d.Set("os", osMap)
		if err != nil {
			return err
		}
	}

	// Set Graphics to resource
	if port != "" {
		graphMap = append(graphMap, map[string]interface{}{
			"listen": listen,
			"port":   port,
			"type":   t,
			"keymap": keymap,
		})
		err = d.Set("graphics", graphMap)
		if err != nil {
			return err
		}
	}

	// Set Disks to Resource
	for _, disk := range vmTemplate.GetDisks() {
		size, _ := disk.GetI(shared.Size)
		driver, _ := disk.Get(shared.Driver)
		target, _ := disk.Get(shared.TargetDisk)
		imageId, _ := disk.GetI(shared.ImageID)

		diskList = append(diskList, map[string]interface{}{
			"image_id": imageId,
			"size":     size,
			"target":   target,
			"driver":   driver,
		})
	}

	if len(diskList) > 0 {
		err = d.Set("disk", diskList)
		if err != nil {
			return err
		}
	}

	// Set Nics to resource
	for i, nic := range vmTemplate.GetNICs() {
		sg := make([]int, 0)
		ip, _ := nic.Get(shared.IP)
		mac, _ := nic.Get(shared.MAC)
		physicalDevice, _ := nic.GetStr("PHYDEV")
		network, _ := nic.Get(shared.Network)
		nicId, _ := nic.ID()

		model, _ := nic.Get(shared.Model)
		networkId, _ := nic.GetI(shared.NetworkID)
		securityGroupsArray, _ := nic.Get(shared.SecurityGroups)

		sgString := strings.Split(securityGroupsArray, ",")
		for _, s := range sgString {
			sgInt, _ := strconv.ParseInt(s, 10, 32)
			sg = append(sg, int(sgInt))
		}

		nicList = append(nicList, map[string]interface{}{
			"ip":              ip,
			"mac":             mac,
			"network_id":      networkId,
			"physical_device": physicalDevice,
			"network":         network,
			"nic_id":          nicId,
			"model":           model,
			"security_groups": sg,
		})
		if i == 0 {
			d.Set("ip", ip)
		}
	}

	if len(nicList) > 0 {
		err = d.Set("nic", nicList)
		if err != nil {
			return err
		}
	}

	if tplTags {
		tags := make(map[string]interface{})
		// Get only tags from userTemplate
		if tagsInterface, ok := d.GetOk("tags"); ok {
			for k, _ := range tagsInterface.(map[string]interface{}) {
				tags[k], err = vmTemplate.GetStr(strings.ToUpper(k))
				if err != nil {
					return err
				}
			}
		}

		if len(tags) > 0 {
			err := d.Set("tags", tags)
			if err != nil {
				return err
			}
		}
	}

	if vmcontext != nil {
		for _, p := range vmcontext.Pairs {
			// Get only contexts elements from VM template
			usercontext := d.Get("context").(map[string]interface{})
			for k, _ := range usercontext {
				if strings.ToUpper(k) == p.Key() {
					context[strings.ToUpper(k)] = p.Value
				}
			}

			if len(context) > 0 {
				err := d.Set("context", context)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
