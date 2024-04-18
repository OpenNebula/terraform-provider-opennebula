package opennebula

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
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
				Default:     defaultVMTimeoutMin,
				Description: "Timeout (in minutes) within resource should be available. Default: 3 minutes",
				Deprecated:  "Native terraform timeout facilities should be used instead",
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
					if !contains(value, vmDiskOnChangeValues) {
						errors = append(errors, fmt.Errorf("%q must be one of %s", k, strings.Join(vmDiskOnChangeValues, ", ")))
					}
					return
				},
			},
			"template_disk": templateDiskVMSchema(),
			"disk":          diskVMSchema(),
			"hard_shutdown": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Immediately poweroff/terminate/reboot/undeploy the VM. (default: false)",
			},
			"template_tags": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "When template_id was set this keeps the template tags.",
			},
			"template_section_names": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "When template_id was set this keeps the template sections names.",
			},
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
		"raw": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Low-level hypervisor tuning",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
							validtypes := []string{"kvm", "lxd", "vmware"}
							value := v.(string)

							if !contains(value, validtypes) {
								errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(validtypes, ",")))
							}

							return
						},
						Description: "Name of the hypervisor: kvm, lxd, vmware",
					},
					"data": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Low-level data to pass to the hypervisor",
					},
				},
			},
		},
		"tags":         tagsSchema(),
		"default_tags": defaultTagsSchemaComputed(),
		"tags_all":     tagsSchemaComputed(),
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
		"template_section":      templateSectionSchema(),
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
			Optional: true,
			Default:  -1,
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
		"network_mode_auto": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"sched_requirements": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"sched_rank": {
			Type:     schema.TypeString,
			Optional: true,
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
		"dev_prefix": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"cache": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"discard": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"io": {
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

				if !contains(value, validtypes) {
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

				if !contains(value, validtypes) {
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
				"passwd": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"graphics.0.random_passwd"},
				},
				"random_passwd": {
					Type:          schema.TypeBool,
					Optional:      true,
					ConflictsWith: []string{"graphics.0.passwd"},
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

var locktypes = []string{"USE", "MANAGE", "ADMIN", "ALL", "UNLOCK"}

func lockSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Lock level of the new resource: USE, MANAGE, ADMIN, ALL, UNLOCK",
		ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
			value := v.(string)

			if !contains(value, locktypes) {
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
		case "dev_prefix":
			disk.Add("DEV_PREFIX", v.(string))
		case "cache":
			disk.Add("CACHE", v.(string))
		case "discard":
			disk.Add("DISCARD", v.(string))
		case "io":
			disk.Add("IO", v.(string))
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
			networkID := v.(int)
			if networkID != -1 {
				nic.Add(shared.NetworkID, strconv.Itoa(networkID))
			}
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
		case "network_mode_auto":
			if v.(bool) {
				nic.Add(shared.NetworkMode, "auto")
			}
		case "sched_requirements":
			nic.Add(shared.SchedRequirements, v.(string))
		case "sched_rank":
			nic.Add(shared.SchedRank, v.(string))

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
			case "passwd":
				tpl.AddIOGraphic(vmk.Passwd, v.(string))
			case "random_passwd":
				// Convert bool to string
				tpl.AddIOGraphic(vmk.RandomPassword, map[bool]string{true: "YES", false: "NO"}[v.(bool)])
			}

		}

	}
}

func addDisks(d *schema.ResourceData, tpl *vm.Template) error {

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

	return nil
}

func generateVMTemplate(d *schema.ResourceData, tpl *vm.Template) error {

	err := addDisks(d, tpl)
	if err != nil {
		return err
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

	vectorsInterface := d.Get("template_section").(*schema.Set).List()
	if len(vectorsInterface) > 0 {
		addTemplateVectors(vectorsInterface, &tpl.Template)
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

	//Generate RAW definition
	raw := d.Get("raw").([]interface{})
	for i := 0; i < len(raw); i++ {
		rawConfig := raw[i].(map[string]interface{})
		rawVec := tpl.AddVector("RAW")
		rawVec.AddPair("TYPE", rawConfig["type"].(string))
		rawVec.AddPair("DATA", rawConfig["data"].(string))
	}

	descr, ok := d.GetOk("description")
	if ok {
		tpl.Add(vmk.Description, descr.(string))
	}

	return nil
}

func updateTemplate(d *schema.ResourceData, tpl *vm.Template) bool {

	update := false

	if d.HasChange("sched_requirements") {
		schedRequirements := d.Get("sched_requirements").(string)

		if len(schedRequirements) > 0 {
			tpl.Placement(vmk.SchedRequirements, schedRequirements)
		} else {
			tpl.Del(string(vmk.SchedRequirements))
		}
		update = true
	}

	if d.HasChange("sched_ds_requirements") {
		schedDSRequirements := d.Get("sched_ds_requirements").(string)

		if len(schedDSRequirements) > 0 {
			tpl.Placement(vmk.SchedDSRequirements, schedDSRequirements)
		} else {
			tpl.Del(string(vmk.SchedDSRequirements))
		}
		update = true
	}

	if d.HasChange("description") {

		tpl.Del(string(vmk.Description))

		description := d.Get("description").(string)

		if len(description) > 0 {
			tpl.Add(vmk.Description, description)
		}

		update = true
	}

	if d.HasChange("template_section") {
		updateTemplateSection(d, &tpl.Template)
		update = true
	}

	return update
}

func updateRaw(d *schema.ResourceData, tpl *dyn.Template) {
	tpl.Del("RAW")

	raw := d.Get("raw").([]interface{})
	if len(raw) > 0 {
		for i := 0; i < len(raw); i++ {
			rawConfig := raw[i].(map[string]interface{})
			rawVec := tpl.AddVector("RAW")
			rawVec.AddPair("TYPE", rawConfig["type"].(string))
			rawVec.AddPair("DATA", rawConfig["data"].(string))
		}
	}
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

	networkModeBool := false
	networkMode, err := nic.Get(shared.NetworkMode)
	if err == nil && networkMode == "auto" {
		networkModeBool = true
	}

	schedReqs, _ := nic.Get(shared.SchedRequirements)
	schedRank, _ := nic.Get(shared.SchedRank)

	securityGroupsArray, _ := nic.Get(shared.SecurityGroups)
	if len(securityGroupsArray) > 0 {
		sgString := strings.Split(securityGroupsArray, ",")
		for _, s := range sgString {
			sgInt, _ := strconv.ParseInt(s, 10, 32)
			sg = append(sg, int(sgInt))
		}
	}

	return map[string]interface{}{
		"ip":                 ip,
		"mac":                mac,
		"network_id":         networkId,
		"physical_device":    physicalDevice,
		"network":            network,
		"model":              model,
		"virtio_queues":      virtioQueues,
		"security_groups":    sg,
		"network_mode_auto":  networkModeBool,
		"sched_requirements": schedReqs,
		"sched_rank":         schedRank,
	}
}

func flattenDisk(disk shared.Disk) map[string]interface{} {

	size, _ := disk.GetI(shared.Size)
	if size == -1 {
		size = 0
	}
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	dev_prefix, _ := disk.Get("DEV_PREFIX")
	cache, _ := disk.Get("CACHE")
	discard, _ := disk.Get("DISCARD")
	io, _ := disk.Get("IO")
	imageID, _ := disk.GetI(shared.ImageID)
	volatileType, _ := disk.Get("TYPE")
	volatileFormat, _ := disk.Get("FORMAT")

	return map[string]interface{}{
		"image_id":        imageID,
		"size":            size,
		"target":          target,
		"dev_prefix":      dev_prefix,
		"cache":           cache,
		"discard":         discard,
		"io":              io,
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

func flattenTemplate(d *schema.ResourceData, inheritedVectors map[string]interface{}, vmTemplate *vm.Template) error {

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
	randomPasswd, _ := vmTemplate.GetIOGraphic(vmk.RandomPassword)

	// Raw
	rawVec, _ := vmTemplate.GetVector("RAW")

	// VM size
	cpu, _ := vmTemplate.GetCPU()
	vcpu, _ := vmTemplate.GetVCPU()
	memory, _ := vmTemplate.GetMemory()

	// Set VM size
	err = d.Set("cpu", cpu)
	if err != nil {
		return err
	}

	err = d.Set("vcpu", vcpu)
	if err != nil {
		return err
	}

	err = d.Set("memory", memory)
	if err != nil {
		return err
	}

	// Set CPU Model to resource
	if cpumodel != "" {
		cpumodelMap = append(cpumodelMap, map[string]interface{}{
			"model": cpumodel,
		})
		_, inherited := inheritedVectors["CPU_MODEL"]
		if !inherited {
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
		_, inherited := inheritedVectors["OS"]
		if !inherited {
			err = d.Set("os", osMap)
			if err != nil {
				return err
			}
		}
	}

	// Set graphics to resource
	if port != "" {
		graphMap = append(graphMap, map[string]interface{}{
			"listen":        listen,
			"port":          port,
			"type":          t,
			"keymap":        keymap,
			"random_passwd": randomPasswd == "YES",
		})
		_, inherited := inheritedVectors["GRAPHICS"]
		if !inherited {
			err = d.Set("graphics", graphMap)
			if err != nil {
				return err
			}
		}
	}

	if rawVec != nil {

		rawMap := make([]map[string]interface{}, 0, 1)

		hypType, _ := rawVec.GetStr("TYPE")
		data, _ := rawVec.GetStr("DATA")

		rawMap = append(rawMap, map[string]interface{}{
			"type": hypType,
			"data": data,
		})

		if _, ok := d.GetOk("raw"); ok {
			err = d.Set("raw", rawMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func flattenVMUserTemplate(d *schema.ResourceData, meta interface{}, inheritedTags map[string]interface{}, vmTemplate *dynamic.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	// We read attributes only if they are described in the VM description
	// to avoid a diff due to template attribute inheritance

	err := flattenTemplateSection(d, meta, vmTemplate)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to read template section",
		})
	}
	tagsDiags := flattenTemplateTags(d, meta, vmTemplate)
	if len(tagsDiags) > 0 {
		diags = append(diags, tagsDiags...)
	}

	schedReq, _ := vmTemplate.GetStr("SCHED_REQUIREMENTS")
	_, inherited := inheritedTags["SCHED_REQUIREMENTS"]
	if !inherited {
		err = d.Set("sched_requirements", schedReq)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to find SCHED_REQUIREMENTS attribute",
		})
	}

	schedDSReq, _ := vmTemplate.GetStr("SCHED_DS_REQUIREMENTS")
	_, inherited = inheritedTags["SCHED_DS_REQUIREMENTS"]
	if !inherited {
		err = d.Set("sched_ds_requirements", schedDSReq)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to find SCHED_DS_REQUIREMENTS attribute",
		})
	}

	description, _ := vmTemplate.GetStr("DESCRIPTION")
	_, inherited = inheritedTags["DESCRIPTION"]
	if !inherited {
		err = d.Set("description", description)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to find DESCRIPTION attribute",
		})
	}

	return nil
}
