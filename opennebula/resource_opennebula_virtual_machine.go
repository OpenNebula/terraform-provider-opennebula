package opennebula

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

var (
	vmDiskUpdateReadyStates = []string{"RUNNING", "POWEROFF"}
	vmDiskResizeReadyStates = []string{"RUNNING", "POWEROFF", "UNDEPLOYED"}
	vmNICUpdateReadyStates  = vmDiskUpdateReadyStates
)

type flattenVMPart func(d *schema.ResourceData, vmTemplate *vm.Template) error

func resourceOpennebulaVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Create:        resourceOpennebulaVirtualMachineCreate,
		Read:          resourceOpennebulaVirtualMachineRead,
		Exists:        resourceOpennebulaVirtualMachineExists,
		Update:        resourceOpennebulaVirtualMachineUpdate,
		Delete:        resourceOpennebulaVirtualMachineDelete,
		CustomizeDiff: resourceVMCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the VM. If empty, defaults to 'templatename-<vmid>'",
			},
			"instance": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Final name of the VM instance",
				Deprecated:  "use 'name' instead",
			},
			"template_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				ForceNew:    true,
				Description: "Id of the VM template to use. Defaults to -1: no template used.",
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
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the template (in Unix format, owner-group-other, use-manage-admin)",
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
				Description: "ID of the user that will own the VM",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the VM",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the VM",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the VM",
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
			"cpu":           cpuSchema(),
			"vcpu":          vcpuSchema(),
			"memory":        memorySchema(),
			"context":       contextSchema(),
			"cpumodel":      cpumodelSchema(),
			"disk":          diskVMSchema(),
			"template_disk": templateDiskVMSchema(),
			"graphics":      graphicsSchema(),
			"nic":           nicVMSchema(),
			"template_nic":  templateNICVMSchema(),
			"os":            osSchema(),
			"vmgroup":       vmGroupSchema(),
			"tags":          tagsSchema(),
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Primary IP address assigned by OpenNebula",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the VM, If empty, it uses caller group",
			},
		},
	}
}

func nicComputedVMFields() map[string]*schema.Schema {

	return map[string]*schema.Schema{
		"nic_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"computed_ip": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_mac": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_model": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_virtio_queues": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_physical_device": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_security_groups": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeInt,
			},
		},
	}

}

func templateNICVMSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Network adapter(s) assigned to the Virtual Machine via a template",
		Elem: &schema.Resource{
			Schema: mergeSchemas(nicComputedVMFields(), map[string]*schema.Schema{
				"network_id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"network": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}),
		},
	}
}

func nicVMFields() map[string]*schema.Schema {
	return mergeSchemas(nicFields(), nicComputedVMFields())
}

func nicVMSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Definition of network adapter(s) assigned to the Virtual Machine",
		Elem: &schema.Resource{
			Schema: nicVMFields(),
		},
	}
}

func diskComputedVMFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"disk_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"computed_size": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"computed_target": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_driver": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func templateDiskVMSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Disks assigned to the Virtual Machine via a template",
		Elem: &schema.Resource{
			Schema: mergeSchemas(diskComputedVMFields(), map[string]*schema.Schema{
				"image_id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			}),
		},
	}
}

func diskVMFields() map[string]*schema.Schema {
	return mergeSchemas(diskFields(), diskComputedVMFields())
}

func diskVMSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Definition of disks assigned to the Virtual Machine",
		Elem: &schema.Resource{
			Schema: diskVMFields(),
		},
	}
}

func getVirtualMachineController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.VMController, error) {
	controller := meta.(*goca.Controller)
	var vmc *goca.VMController

	if d.Id() != "" {

		// Try to find the VM by ID, if specified

		id, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		vmc = controller.VM(int(id))

	} else {

		// Try to find the VM by name as the de facto compound primary key

		id, err := controller.VMs().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
		}
		vmc = controller.VM(id)

	}

	return vmc, nil
}

func changeVmGroup(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var gid int

	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		gid, err = controller.Groups().ByName(d.Get("group").(string))
		if err != nil {
			return err
		}
	} else {
		gid = d.Get("gid").(int)
	}

	err = vmc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	//Call one.template.instantiate only if template_id is defined
	//otherwise use one.vm.allocate
	var err error
	var vmID int

	// If template_id is set to -1 it means not template id to instanciate. This is a workaround
	// because GetOk helper from terraform considers 0 as a Zero() value from an integer.
	templateID := d.Get("template_id").(int)
	if templateID != -1 {
		// if template id is set, instantiate a VM from this template
		tc := controller.Template(templateID)

		// retrieve the context of the template
		tpl, err := tc.Info(true, false)
		if err != nil {
			return err
		}

		tplContext, _ := tpl.Template.GetVector(vmk.ContextVec)

		// customize template except for memory and cpu.
		vmDef, err := generateVm(d, tplContext)
		if err != nil {
			return err
		}

		// Instantiate template without creating a persistent copy of the template
		// Note that the new VM is not pending
		vmID, err = tc.Instantiate(d.Get("name").(string), d.Get("pending").(bool), vmDef, false)
		if err != nil {
			return err
		}
	} else {
		if _, ok := d.GetOk("cpu"); !ok {
			return fmt.Errorf("cpu is mandatory as template_id is not used")
		}
		if _, ok := d.GetOk("memory"); !ok {
			return fmt.Errorf("memory is mandatory as template_id is not used")
		}

		vmDef, err := generateVm(d, nil)
		if err != nil {
			return err
		}

		// Create VM not in pending state
		vmID, err = controller.VMs().Create(vmDef, d.Get("pending").(bool))
		if err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprintf("%v", vmID))
	vmc := controller.VM(vmID)

	expectedState := "RUNNING"
	if d.Get("pending").(bool) {
		expectedState = "HOLD"
	}

	timeout := d.Get("timeout").(int)
	_, err = waitForVMState(vmc, timeout, expectedState)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for virtual machine (%s) to be in state %s: %s", d.Id(), expectedState, err)

	}

	//Set the permissions on the VM if it was defined, otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = vmc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			log.Printf("[ERROR] template permissions change failed, error: %s", err)
			return err
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err = changeVmGroup(d, meta)
		if err != nil {
			return err
		}
	}

	// Customize read step to process disk and NIC from template in a different way.
	// The goal is to avoid diffs that would trigger unwanted disk/NIC update.
	if templateID != -1 {

		flattenDiskFunc := flattenVMDisk
		flattenNICFunc := flattenVMNIC

		if len(d.Get("disk").([]interface{})) == 0 {
			// if no disks overrides those from templates
			flattenDiskFunc = flattenVMTemplateDisk
		} else {
			d.Set("template_disk", []interface{}{})
		}

		if len(d.Get("nic").([]interface{})) == 0 {
			// if no nics overrides those from templates
			flattenNICFunc = flattenVMTemplateNIC
		} else {
			d.Set("template_nic", []interface{}{})
		}

		return resourceOpennebulaVirtualMachineReadCustom(d, meta, flattenDiskFunc, flattenNICFunc)

	}

	d.Set("template_nic", []interface{}{})
	d.Set("template_disk", []interface{}{})

	return resourceOpennebulaVirtualMachineRead(d, meta)
}

func resourceOpennebulaVirtualMachineReadCustom(d *schema.ResourceData, meta interface{}, flattenVMDisk, flattenVMNIC flattenVMPart) error {
	vmc, err := getVirtualMachineController(d, meta, -2, -1, -1)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual machine %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		return err
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", vm.ID))
	d.Set("name", vm.Name)
	d.Set("uid", vm.UID)
	d.Set("gid", vm.GID)
	d.Set("uname", vm.UName)
	d.Set("gname", vm.GName)
	d.Set("state", vm.StateRaw)
	d.Set("lcmstate", vm.LCMStateRaw)
	//TODO fix this:
	err = d.Set("permissions", permissionsUnixString(*vm.Permissions))
	if err != nil {
		return err
	}

	err = flattenVMDisk(d, &vm.Template)
	if err != nil {
		return err
	}

	err = flattenVMNIC(d, &vm.Template)
	if err != nil {
		return err
	}

	err = flattenTemplate(d, &vm.Template, false)
	if err != nil {
		return err
	}

	err = flattenTags(d, &vm.UserTemplate)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	return resourceOpennebulaVirtualMachineReadCustom(d, meta, flattenVMDisk, flattenVMNIC)
}

func flattenDiskComputed(disk shared.Disk) map[string]interface{} {
	size, _ := disk.GetI(shared.Size)
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	diskID, _ := disk.GetI(shared.DiskID)

	return map[string]interface{}{
		"disk_id":         diskID,
		"computed_size":   size,
		"computed_target": target,
		"computed_driver": driver,
	}
}

// flattenVMTemplateDisk read disk that come from template when instantiating a VM
func flattenVMTemplateDisk(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set disks to resource
	disks := vmTemplate.GetDisks()
	diskList := make([]interface{}, 0, len(disks))

	for _, disk := range disks {

		imageID, _ := disk.GetI(shared.ImageID)
		diskRead := flattenDiskComputed(disk)
		diskRead["image_id"] = imageID
		diskList = append(diskList, diskRead)
	}

	err := d.Set("template_disk", diskList)
	if err != nil {
		return err
	}

	return nil
}

// flattenVMDisk is similar to flattenDisk but deal with computed_* attributes
// this is a temporary solution until we can use nested attributes marked computed and optional
func flattenVMDisk(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set disks to resource
	disks := vmTemplate.GetDisks()
	diskList := make([]interface{}, 0, len(disks))

diskLoop:
	for _, disk := range disks {

		imageID, _ := disk.GetI(shared.ImageID)

		// exclude disk from template_disk based on the image_id
		tplDiskConfigs := d.Get("template_disk").([]interface{})
		for _, tplDiskConfigIf := range tplDiskConfigs {
			tplDiskConfig := tplDiskConfigIf.(map[string]interface{})

			if tplDiskConfig["image_id"] == imageID {
				continue diskLoop
			}
		}

		diskRead := flattenDiskComputed(disk)
		diskRead["image_id"] = imageID

		// copy disk config values
		diskConfigs := d.Get("disk").([]interface{})
		for j := 0; j < len(diskConfigs); j++ {
			diskConfig := diskConfigs[j].(map[string]interface{})

			if diskConfig["image_id"] != diskRead["image_id"] {
				continue
			}

			diskRead["size"] = diskConfig["size"]
			diskRead["target"] = diskConfig["target"]
			diskRead["driver"] = diskConfig["driver"]
			break

		}

		diskList = append(diskList, diskRead)
	}

	if len(diskList) > 0 {
		err := d.Set("disk", diskList)
		if err != nil {
			return err
		}
	}

	return nil
}

func flattendNICComputed(nic shared.NIC) map[string]interface{} {
	nicID, _ := nic.ID()
	sg := make([]int, 0)
	ip, _ := nic.Get(shared.IP)
	mac, _ := nic.Get(shared.MAC)
	physicalDevice, _ := nic.GetStr("PHYDEV")
	network, _ := nic.Get(shared.Network)

	model, _ := nic.Get(shared.Model)
	virtioQueues, _ := nic.GetStr("VIRTIO_QUEUES")
	securityGroupsArray, _ := nic.Get(shared.SecurityGroups)

	sgString := strings.Split(securityGroupsArray, ",")
	for _, s := range sgString {
		sgInt, _ := strconv.ParseInt(s, 10, 32)
		sg = append(sg, int(sgInt))
	}

	return map[string]interface{}{
		"nic_id":                   nicID,
		"network":                  network,
		"computed_ip":              ip,
		"computed_mac":             mac,
		"computed_physical_device": physicalDevice,
		"computed_model":           model,
		"computed_virtio_queues":   virtioQueues,
		"computed_security_groups": sg,
	}
}

// flattenVMTemplateNIC read NIC that come from template when instantiating a VM
func flattenVMTemplateNIC(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set Nics to resource
	nics := vmTemplate.GetNICs()
	nicList := make([]interface{}, 0, len(nics))

	for i, nic := range nics {

		networkID, _ := nic.GetI(shared.NetworkID)
		nicRead := flattendNICComputed(nic)
		nicRead["network_id"] = networkID
		nicList = append(nicList, nicRead)

		if i == 0 {
			d.Set("ip", nicRead["computed_ip"])
		}
	}

	err := d.Set("template_nic", nicList)
	if err != nil {
		return err
	}

	return nil
}

// flattenVMNIC is similar to flattenNIC but deal with computed_* attributes
// this is a temporary solution until we can use nested attributes marked computed and optional
func flattenVMNIC(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set Nics to resource
	nics := vmTemplate.GetNICs()
	nicList := make([]interface{}, 0, len(nics))

NICLoop:
	for i, nic := range nics {

		networkID, _ := nic.GetI(shared.NetworkID)

		// exclude NIC from template_nic based on the network_id
		tplNICConfigs := d.Get("template_nic").([]interface{})
		for _, tplNICConfigIf := range tplNICConfigs {
			tplNICConfig := tplNICConfigIf.(map[string]interface{})

			if tplNICConfig["network_id"] == networkID {
				continue NICLoop
			}
		}

		nicRead := flattendNICComputed(nic)
		nicRead["network_id"] = networkID

		// copy nic config values
		nicsConfigs := d.Get("nic").([]interface{})
		for j := 0; j < len(nicsConfigs); j++ {
			nicConfig := nicsConfigs[j].(map[string]interface{})

			if nicConfig["network_id"] != nicRead["network_id"] {
				continue
			}

			nicRead["ip"] = nicConfig["ip"]
			nicRead["mac"] = nicConfig["mac"]
			nicRead["model"] = nicConfig["model"]
			nicRead["virtio_queues"] = nicConfig["virtio_queues"]
			nicRead["physical_device"] = nicConfig["physical_device"]
			nicRead["security_groups"] = nicConfig["security_groups"]

			break

		}

		nicList = append(nicList, nicRead)

		if i == 0 {
			d.Set("ip", nicRead["computed_ip"])
		}
	}

	if len(nicList) > 0 {
		err := d.Set("nic", nicList)
		if err != nil {
			return err
		}
	}
	return nil
}

func flattenTags(d *schema.ResourceData, vmUserTpl *vm.UserTemplate) error {

	tags := make(map[string]interface{})
	for i, _ := range vmUserTpl.Elements {
		pair, ok := vmUserTpl.Elements[i].(*dyn.Pair)
		if !ok {
			continue
		}

		// Get only tags from userTemplate
		tagsInterface := d.Get("tags").(map[string]interface{})
		for k, _ := range tagsInterface {
			if strings.ToUpper(k) == pair.Key() {
				tags[k] = pair.Value
			}
		}
	}

	if len(tags) > 0 {
		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceOpennebulaVirtualMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaVirtualMachineRead(d, meta)
	// a terminated VM is in state 6 (DONE)
	if err != nil || d.Id() == "" || d.Get("state").(int) == 6 {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaVirtualMachineUpdate(d *schema.ResourceData, meta interface{}) error {

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		err := vmc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		vm, err := vmc.Info(false)
		log.Printf("[INFO] Successfully updated name (%s) for VM ID %x\n", vm.Name, vm.ID)
	}

	if d.HasChange("permissions") && d.Get("permissions") != "" {
		if perms, ok := d.GetOk("permissions"); ok {
			err = vmc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated Permissions VM %s\n", vm.Name)
	}

	if d.HasChange("group") {
		err := changeVmGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for VM %s\n", vm.Name)
	}

	if d.HasChange("tags") {
		tagsInterface := d.Get("tags").(map[string]interface{})
		for k, v := range tagsInterface {
			vm.UserTemplate.Del(strings.ToUpper(k))
			vm.UserTemplate.AddPair(strings.ToUpper(k), v.(string))
		}

		err = vmc.Update(vm.UserTemplate.String(), 1)
		if err != nil {
			return err
		}
	}

	if d.HasChange("disk") {

		log.Printf("[INFO] Update disk configuration")

		old, new := d.GetChange("disk")
		attachedDisksCfg := old.([]interface{})
		newDisksCfg := new.([]interface{})

		timeout := d.Get("timeout").(int)

		// get unique elements of each list of configs
		toDetach, toAttach := diffListConfig(newDisksCfg, attachedDisksCfg,
			&schema.Resource{
				Schema: diskFields(),
			},
			"image_id",
			"target",
			"driver")

		// get disks to resize
		_, toResize := diffListConfig(newDisksCfg, attachedDisksCfg,
			&schema.Resource{
				Schema: diskFields(),
			},
			"size")

		// Detach the disks
		var diskID int
		for _, diskIf := range toDetach {
			diskConfig := diskIf.(map[string]interface{})

			imageID := diskConfig["image_id"].(int)
			if imageID == -1 {
				continue
			}

			// retrieve the the disk_id
			for _, d := range attachedDisksCfg {
				cfg := d.(map[string]interface{})
				if cfg["image_id"].(int) != diskConfig["image_id"].(int) {
					continue
				}
				diskID = cfg["disk_id"].(int)
				break
			}

			err := vmDiskDetach(vmc, timeout, diskID)
			if err != nil {
				return fmt.Errorf("vm disk detach: %s", err)

			}
		}

		// Attach the disks
		for _, diskIf := range toAttach {
			diskConfig := diskIf.(map[string]interface{})

			imageID := diskConfig["image_id"].(int)
			if imageID == -1 {
				continue
			}

			diskTpl := makeDiskVector(diskConfig)

			err := vmDiskAttach(vmc, timeout, diskTpl)
			if err != nil {
				return fmt.Errorf("vm disk attach: %s", err)
			}
		}

		// Resize disks
		for _, diskIf := range toResize {
			diskConfig := diskIf.(map[string]interface{})

			imageID := diskConfig["image_id"].(int)
			if imageID == -1 {
				continue
			}

			// retrieve the the disk_id
			for _, d := range attachedDisksCfg {
				cfg := d.(map[string]interface{})
				if cfg["image_id"].(int) != diskConfig["image_id"].(int) {
					continue
				}
				diskID = cfg["disk_id"].(int)

				if diskConfig["size"].(int) > cfg["computed_size"].(int) {
					err := vmDiskResize(vmc, timeout, diskID, diskConfig["size"].(int))
					if err != nil {
						return fmt.Errorf("vm disk resize: %s", err)
					}
				}
			}
		}
	}

	if d.HasChange("nic") {

		log.Printf("[INFO] Update NIC configuration")

		old, new := d.GetChange("nic")
		attachedNicsCfg := old.([]interface{})
		newNicsCfg := new.([]interface{})

		timeout := d.Get("timeout").(int)

		// get unique elements of each list of configs
		toDetach, toAttach := diffListConfig(newNicsCfg, attachedNicsCfg,
			&schema.Resource{
				Schema: nicFields(),
			},
			"network_id",
			"ip",
			"mac",
			"security_groups",
			"model",
			"virtio_queues",
			"physical_device")

		// Detach the nics
		var nicID int
		for _, nicIf := range toDetach {
			nicConfig := nicIf.(map[string]interface{})

			// retrieve the the nic_id
			for _, d := range attachedNicsCfg {
				cfg := d.(map[string]interface{})
				if cfg["network_id"].(int) != nicConfig["network_id"].(int) {
					continue
				}
				nicID = cfg["nic_id"].(int)
				break
			}

			err := vmNICDetach(vmc, timeout, nicID)
			if err != nil {
				return fmt.Errorf("vm nic detach: %s", err)

			}
		}

		// Attach the nics
		for _, nicIf := range toAttach {
			nicConfig := nicIf.(map[string]interface{})

			nicTpl := makeNICVector(nicConfig)

			err := vmNICAttach(vmc, timeout, nicTpl)
			if err != nil {
				return fmt.Errorf("vm nic attach: %s", err)
			}
		}
	}

	return resourceOpennebulaVirtualMachineRead(d, meta)
}

func resourceOpennebulaVirtualMachineDelete(d *schema.ResourceData, meta interface{}) error {
	err := resourceOpennebulaVirtualMachineRead(d, meta)
	if err != nil || d.Id() == "" {
		return err
	}

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	if err = vmc.TerminateHard(); err != nil {
		return err
	}

	timeout := d.Get("timeout").(int)
	ret, err := waitForVMState(vmc, timeout, "DONE")
	if err != nil {

		log.Printf("[WARN] %s\n", err)

		// Retry if timeout not reached
		_, ok := err.(*resource.TimeoutError)
		if !ok && ret != nil {

			vmInfos, _ := ret.(*vm.VM)
			vmState, vmLcmState, _ := vmInfos.State()
			if vmState == vm.Active && vmLcmState == vm.EpilogFailure {

				log.Printf("[INFO] retry terminate VM\n")

				err := vmc.TerminateHard()
				if err != nil {
					return err
				}

				_, err = waitForVMState(vmc, timeout, "DONE")
				if err != nil {
					return err
				}

			} else {
				return fmt.Errorf(
					"Error waiting for virtual machine (%s) to be in state DONE: %s (state: %v, lcmState: %v)", d.Id(), err, vmState, vmLcmState)
			}
		}
	}

	log.Printf("[INFO] Successfully terminated VM\n")
	return nil
}

func waitForVMState(vmc *goca.VMController, timeout int, states ...string) (interface{}, error) {

	stateConf := &resource.StateChangeConf{
		Pending: []string{"anythingelse"},
		Target:  states,
		Refresh: func() (interface{}, string, error) {

			log.Println("Refreshing VM state...")

			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			vmInfos, err := vmc.Info(false)
			if err != nil {
				if NoExists(err) {
					// Do not return an error here as it is excpected if the VM is already in DONE state
					// after its destruction
					return vmInfos, "notfound", nil
				}
				return vmInfos, "", err
			}

			vmState, vmLcmState, err := vmInfos.State()
			if err != nil {
				return vmInfos, "", err
			}
			log.Printf("VM (ID:%d, name:%s) is currently in state %s and in LCM state %s", vmInfos.ID, vmInfos.Name, vmState.String(), vmLcmState.String())

			switch vmState {

			case vm.Done, vm.Hold:
				return vmInfos, vmState.String(), nil
			case vm.Active:
				switch vmLcmState {
				case vm.Running:
					return vmInfos, vmLcmState.String(), nil
				case vm.BootFailure, vm.PrologFailure, vm.EpilogFailure:
					vmerr, _ := vmInfos.UserTemplate.Get(vmk.Error)
					return vmInfos, vmLcmState.String(), fmt.Errorf("VM (ID:%d) entered fail state, error: %s", vmInfos.ID, vmerr)
				default:
					return vmInfos, "anythingelse", nil
				}
			default:
				return vmInfos, "anythingelse", nil
			}

		},
		Timeout:    time.Duration(timeout) * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func generateVm(d *schema.ResourceData, tplContext *dyn.Vector) (string, error) {

	tpl := vm.NewTemplate()

	if d.Get("name") != nil {
		tpl.Add(vmk.Name, d.Get("name").(string))
	}

	//Generate CONTEXT definition
	context := d.Get("context").(map[string]interface{})
	log.Printf("Number of CONTEXT vars: %d", len(context))
	log.Printf("CONTEXT Map: %s", context)

	if tplContext != nil {

		// Update existing context:
		// - add new pairs
		// - update pair when the key already exist
		// - other pairs are left unchanged
		for key, value := range context {
			keyUp := strings.ToUpper(key)
			tplContext.Del(keyUp)
			tplContext.AddPair(keyUp, value)
		}

		tpl.Elements = append(tpl.Elements, tplContext)
	} else {

		// Add new context elements to the template
		for key, value := range context {
			keyUp := strings.ToUpper(key)
			tpl.AddCtx(vmk.Context(keyUp), fmt.Sprint(value))
		}
	}

	generateVMTemplate(d, tpl)

	tplStr := tpl.String()
	log.Printf("[INFO] VM definition: %s", tplStr)

	return tplStr, nil
}

func resourceVMCustomizeDiff(diff *schema.ResourceDiff, v interface{}) error {
	// If the VM is in error state, force the VM to be recreated
	if diff.Get("lcmstate") == 36 {
		log.Printf("[INFO] VM is in error state, forcing recreate.")
		diff.SetNew("lcmstate", 3)
		if err := diff.ForceNew("lcmstate"); err != nil {
			return err
		}
	}

	return nil
}
