package opennebula

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

func commonVMSchemas() map[string]*schema.Schema {
	return mergeSchemas(
		commonInstanceSchema(),
		map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the VM. If empty, defaults to 'templatename-<vmid>'",
			},
			"pending": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Pending state of the VM during its creation, by default it is set to false",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "Timeout (in minutes) within resource should be available. Default: 3 minutes",
			},
			"state": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Current state of the VM",
			},
			"lcmstate": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Current LCM state of the VM",
			},
			"on_disk_change": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "swap", //"recreate" or "swap",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := strings.ToUpper(v.(string))
					if inArray(value, vmDiskOnChangeValues) == -1 {
						errors = append(errors, fmt.Errorf("%q must be one of %s", k, strings.Join(vmDiskOnChangeValues, ", ")))
					}
					return
				},
			},
			"template_disk": templateDiskVMSchema(),
			"disk":          diskVMSchema(),
		},
	)
}

func commonInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cpu":      cpuSchema(),
		"vcpu":     vcpuSchema(),
		"memory":   memorySchema(),
		"context":  contextSchema(),
		"cpumodel": cpumodelSchema(),
		"graphics": graphicsSchema(),
		"os":       osSchema(),
		"vmgroup":  vmGroupSchema(),
		"tags":     tagsSchema(),
		"permissions": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Permissions for the resource (in Unix format, owner-group-other, use-manage-admin)",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				value := v.(string)

				if len(value) != 3 {
					errors = append(errors, fmt.Errorf("%q has specify 3 permission sets: owner-group-other", k))
				}

				all := true
				for _, c := range strings.Split(value, "") {
					if c < "0" || c > "7" {
						all = false
					}
				}
				if !all {
					errors = append(errors, fmt.Errorf("Each character in %q should specify a Unix-like permission set with a number from 0 to 7", k))
				}

				return
			},
		},
		"uid": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "ID of the user that will own the resource",
		},
		"gid": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "ID of the group that will own the resource",
		},
		"uname": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the user that will own the resource",
		},
		"gname": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the group that will own the resource",
		},
		"group": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the Group that onws the resource, If empty, it uses caller group",
		},
		"lock":                  lockSchema(),
		"sched_requirements":    schedReqSchema(),
		"sched_ds_requirements": schedDSReqSchema(),
		"description":           descriptionSchema(),
	}
}

func nicFields() map[string]*schema.Schema {

	return map[string]*schema.Schema{
		"ip": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"mac": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"model": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"virtio_queues": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Only if model is virtio",
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
			Optional: true,
		},
		"security_groups": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
		},
	}
}

func nicSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Definition of network adapter(s) assigned to the Virtual Machine",
		Elem: &schema.Resource{
			Schema: nicFields(),
		},
	}
}

func diskFields(customFields ...map[string]*schema.Schema) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"image_id": {
			Type:        schema.TypeInt,
			Default:     -1,
			Optional:    true,
			Description: "Image Id  of the image to attach to the VM. Defaults to -1: no image attached.",
		},
		"size": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"target": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"driver": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"volatile_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Type of the volatile disk: swap or fs.",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				validtypes := []string{"swap", "fs"}
				value := v.(string)

				if inArray(value, validtypes) < 0 {
					errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(validtypes, ",")))
				}

				return
			},
		},
		"volatile_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Format of the volatile disk: raw or qcow2.",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				validtypes := []string{"raw", "qcow2"}
				value := v.(string)

				if inArray(value, validtypes) < 0 {
					errors = append(errors, fmt.Errorf("Format %q must be one of: %s", k, strings.Join(validtypes, ",")))
				}

				return
			},
		},
	}
}

func diskSchema(customFields ...map[string]*schema.Schema) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Definition of disks assigned to the Virtual Machine",
		Elem: &schema.Resource{
			Schema: diskFields(),
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

func cpumodelSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Definition of CPU Model type for the Virtual Machine",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"model": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func graphicsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
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
		Description: "Add custom tags to the resource",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

var locktypes = []string{"USE", "MANAGE", "ADMIN", "ALL", "UNLOCK"}

func lockSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Lock level of the new resource: USE, MANAGE, ADMIN, ALL, UNLOCK",
		ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
			value := v.(string)

			if inArray(value, locktypes) < 0 {
				errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(locktypes, ",")))
			}

			return
		},
	}
}

func schedReqSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Scheduling requirements to deploy the resource following specific rule",
	}
}

func schedDSReqSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Storage placement requirements to deploy the resource following specific rule",
	}
}

func descriptionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "A description of the entity",
	}
}

func makeDiskVector(diskConfig map[string]interface{}) *shared.Disk {
	disk := shared.NewDisk()

	for k, v := range diskConfig {

		if k == "image_id" {
			imageID := v.(int)
			if imageID >= 0 {
				disk.Add(shared.ImageID, strconv.Itoa(imageID))
			}
			continue
		}

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
		case "volatile_type":
			disk.Add("TYPE", v.(string))
		case "volatile_format":
			disk.Add("FORMAT", v.(string))
		}
	}

	return disk
}

func makeNICVector(nicConfig map[string]interface{}) *shared.NIC {
	nic := shared.NewNIC()

	for k, v := range nicConfig {

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
		case "virtio_queues":
			nic.Add("VIRTIO_QUEUES", v.(string))
		case "physical_device":
			nic.Add("PHYDEV", v.(string))
		case "security_groups":
			nicSecGroups := ArrayToString(v.([]interface{}), ",")
			nic.Add(shared.SecurityGroups, nicSecGroups)
		}
	}

	return nic
}

func addOS(tpl *vm.Template, os []interface{}) {

	for i := 0; i < len(os); i++ {
		osconfig := os[i].(map[string]interface{})
		tpl.AddOS(vmk.Arch, osconfig["arch"].(string))
		tpl.AddOS(vmk.Boot, osconfig["boot"].(string))
	}

}

func addGraphic(tpl *vm.Template, graphics []interface{}) {

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
}

func generateVMTemplate(d *schema.ResourceData, tpl *vm.Template) error {

	//Generate DISK definition
	disks := d.Get("disk").([]interface{})
	log.Printf("Number of disks: %d", len(disks))

	for i := 0; i < len(disks); i++ {
		diskconfig := disks[i].(map[string]interface{})

		// ConflictsWith can't be used among attributes of a nested part: disk, nic etc.
		// So we need to add a check here
		if diskconfig["image_id"].(int) != -1 &&
			(len(diskconfig["volatile_type"].(string)) > 0 ||
				len(diskconfig["volatile_format"].(string)) > 0) {
			return fmt.Errorf("disk attritutes image_id can't be defined at the same time as volatile_type or volatile_format")
		}

		// Ignore disk creation if Image ID is -1
		if diskconfig["image_id"].(int) == -1 &&
			len(diskconfig["volatile_type"].(string)) == 0 {
			continue
		}

		disk := makeDiskVector(diskconfig)
		tpl.Elements = append(tpl.Elements, disk)
	}

	//Generate GRAPHICS definition
	addGraphic(tpl, d.Get("graphics").([]interface{}))

	//Generate OS definition
	addOS(tpl, d.Get("os").([]interface{}))

	//Generate CPU Model definition
	cpumodel := d.Get("cpumodel").([]interface{})
	for i := 0; i < len(cpumodel); i++ {
		cpumodelconfig := cpumodel[i].(map[string]interface{})
		tpl.CPUModel(cpumodelconfig["model"].(string))
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

	schedReq, ok := d.GetOk("sched_requirements")
	if ok {
		tpl.AddPair("SCHED_REQUIREMENTS", schedReq.(string))

	}

	schedDSReq, ok := d.GetOk("sched_ds_requirements")
	if ok {
		tpl.AddPair("SCHED_DS_REQUIREMENTS", schedDSReq.(string))

	}

	descr, ok := d.GetOk("description")
	if ok {
		tpl.Add(vmk.Description, descr.(string))
	}

	return nil
}

func flattenNIC(nic shared.NIC) map[string]interface{} {

	sg := make([]int, 0)
	ip, _ := nic.Get(shared.IP)
	mac, _ := nic.Get(shared.MAC)
	physicalDevice, _ := nic.GetStr("PHYDEV")
	network, _ := nic.Get(shared.Network)

	model, _ := nic.Get(shared.Model)
	virtioQueues, _ := nic.GetStr("VIRTIO_QUEUES")
	networkId, _ := nic.GetI(shared.NetworkID)
	securityGroupsArray, _ := nic.Get(shared.SecurityGroups)

	if len(securityGroupsArray) > 0 {
		sgString := strings.Split(securityGroupsArray, ",")
		for _, s := range sgString {
			sgInt, _ := strconv.ParseInt(s, 10, 32)
			sg = append(sg, int(sgInt))
		}
	}

	return map[string]interface{}{
		"ip":              ip,
		"mac":             mac,
		"network_id":      networkId,
		"physical_device": physicalDevice,
		"network":         network,
		"model":           model,
		"virtio_queues":   virtioQueues,
		"security_groups": sg,
	}
}

func flattenDisk(disk shared.Disk) map[string]interface{} {

	size, _ := disk.GetI(shared.Size)
	if size == -1 {
		size = 0
	}
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	imageID, _ := disk.GetI(shared.ImageID)
	volatileType, _ := disk.Get("TYPE")
	volatileFormat, _ := disk.Get("FORMAT")

	return map[string]interface{}{
		"image_id":        imageID,
		"size":            size,
		"target":          target,
		"driver":          driver,
		"volatile_type":   volatileType,
		"volatile_format": volatileFormat,
	}
}

func flattenTemplateVMGroup(d *schema.ResourceData, vmTemplate *vm.Template) error {
	var err error

	// VM Group
	vmgMap := make([]map[string]interface{}, 0, 1)
	vmgIdStr, _ := vmTemplate.GetStrFromVec("VMGROUP", "VMGROUP_ID")
	vmgid, _ := strconv.ParseInt(vmgIdStr, 10, 32)
	vmgRole, _ := vmTemplate.GetStrFromVec("VMGROUP", "ROLE")

	// Set VM Group to resource
	if vmgIdStr != "" {
		vmgMap = append(vmgMap, map[string]interface{}{
			"vmgroup_id": vmgid,
			"role":       vmgRole,
		})
		if _, ok := d.GetOk("vmgroup"); ok {
			err = d.Set("vmgroup", vmgMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func flattenTemplate(d *schema.ResourceData, vmTemplate *vm.Template) error {

	var err error

	// OS
	osMap := make([]map[string]interface{}, 0, 1)
	arch, _ := vmTemplate.GetOS(vmk.Arch)
	boot, _ := vmTemplate.GetOS(vmk.Boot)
	// CPU Model
	cpumodelMap := make([]map[string]interface{}, 0, 1)
	cpumodel, _ := vmTemplate.GetCPUModel(vmk.Model)
	// Graphics
	graphMap := make([]map[string]interface{}, 0, 1)
	listen, _ := vmTemplate.GetIOGraphic(vmk.Listen)
	port, _ := vmTemplate.GetIOGraphic(vmk.Port)
	t, _ := vmTemplate.GetIOGraphic(vmk.GraphicType)
	keymap, _ := vmTemplate.GetIOGraphic(vmk.Keymap)
	// Features
	featuresMap := make([]map[string]interface{}, 0, 1)
	pae, _ := vmTemplate.GetFeature(vmk.PAE)
	acpi, _ := vmTemplate.GetFeature(vmk.ACPI)
	apic, _ := vmTemplate.GetFeature(vmk.APIC)
	localtime, _ := vmTemplate.GetFeature(vmk.LocalTime)
	// not using vmk here because key not defined yet:
	hyperv, _ := vmTemplate.GetFeature("HYPERV")
	guest_agent, _ := vmTemplate.GetFeature(vmk.GuestAgent)
	virtio_scsi_queues, _ := vmTemplate.GetFeature(vmk.VirtIOScsiQueues)
	// not using vmk here because key not defined yet:
	iothreads, _ := vmTemplate.GetFeature("IOTHREADS")

	// Set CPU Model to resource
	if cpumodel != "" {
		cpumodelMap = append(cpumodelMap, map[string]interface{}{
			"model": cpumodel,
		})
		if _, ok := d.GetOk("cpumodel"); ok {
			err = d.Set("cpumodel", cpumodelMap)
			if err != nil {
				return err
			}
		}
	}

	err = flattenTemplateVMGroup(d, vmTemplate)
	if err != nil {
		return err
	}

	// Set OS to resource
	if arch != "" {
		osMap = append(osMap, map[string]interface{}{
			"arch": arch,
			"boot": boot,
		})
		if _, ok := d.GetOk("os"); ok {
			err = d.Set("os", osMap)
			if err != nil {
				return err
			}
		}
	}

	// Set graphics to resource
	if port != "" {
		graphMap = append(graphMap, map[string]interface{}{
			"listen": listen,
			"port":   port,
			"type":   t,
			"keymap": keymap,
		})
		if _, ok := d.GetOk("graphics"); ok {
			err = d.Set("graphics", graphMap)
			if err != nil {
				return err
			}
		}
	}

	// Set features to resource
	if pae != "" || acpi != "" || apic != "" || localtime != "" || hyperv != "" || guest_agent != "" || virtio_scsi_queues != "" || iothreads != "" {
		featuresMap = append(featuresMap, map[string]interface{}{
			"pae":                pae,
			"acpi":               acpi,
			"apic":               apic,
			"localtime":          localtime,
			"hyperv":             hyperv,
			"guest_agent":        guest_agent,
			"virtio_scsi_queues": virtio_scsi_queues,
			"iothreads":          iothreads,
		})
		if _, ok := d.GetOk("features"); ok {
			err = d.Set("features", featuresMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func flattenUserTemplate(d *schema.ResourceData, vmTemplate *dynamic.Template) error {

	var err error

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

	schedReq, _ := vmTemplate.GetStr("SCHED_REQUIREMENTS")
	if len(schedReq) > 0 {
		err = d.Set("sched_requirements", schedReq)
		if err != nil {
			return err
		}
	}

	schedDSReq, _ := vmTemplate.GetStr("SCHED_DS_REQUIREMENTS")
	if len(schedDSReq) > 0 {
		err = d.Set("sched_ds_requirements", schedDSReq)
		if err != nil {
			return err
		}
	}

	desc, _ := vmTemplate.GetStr("DESCRIPTION")
	if len(desc) > 0 {
		err = d.Set("description", desc)
		if err != nil {
			return err
		}
	}

	return nil
}
