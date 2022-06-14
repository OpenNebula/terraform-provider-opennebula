package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

var (
	vmDiskOnChangeValues    = []string{"RECREATE", "SWAP"}
	vmDiskUpdateReadyStates = []string{"RUNNING", "POWEROFF"}
	vmDiskResizeReadyStates = []string{"RUNNING", "POWEROFF", "UNDEPLOYED"}
	vmNICUpdateReadyStates  = vmDiskUpdateReadyStates
	vmDeleteReadyStates     = []string{"RUNNING", "HOLD", "POWEROFF", "STOPPED", "UNDEPLOYED", "SUSPENDED"}
	defaultVMMinTimeout     = 3
	defaultVMTimeout        = time.Duration(defaultVMMinTimeout) * time.Minute
)

func resourceOpennebulaVirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualMachineCreate,
		ReadContext:   resourceOpennebulaVirtualMachineRead,
		Exists:        resourceOpennebulaVirtualMachineExists,
		UpdateContext: resourceOpennebulaVirtualMachineUpdate,
		DeleteContext: resourceOpennebulaVirtualMachineDelete,
		CustomizeDiff: resourceVMCustomizeDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultVMTimeout),
			Update: schema.DefaultTimeout(defaultVMTimeout),
			Delete: schema.DefaultTimeout(defaultVMTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: mergeSchemas(
			commonVMSchemas(),
			map[string]*schema.Schema{
				"instance": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Final name of the VM instance",
					Deprecated:  "use 'name' instead",
				},
				"ip": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Primary IP address assigned by OpenNebula",
				},
				"nic": nicVMSchema(),
				"keep_nic_order": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Force the provider to keep nics order at update.",
				},
				"template_id": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     -1,
					ForceNew:    true,
					Description: "Id of the VM template to use. Defaults to -1: no template used.",
				},
				"template_nic": templateNICVMSchema(),
			},
		),
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
		"computed_volatile_format": {
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
	config := meta.(*Configuration)
	controller := config.Controller
	var vmc *goca.VMController

	if d.Id() != "" {

		// Try to find the VM by ID, if specified

		id, err := strconv.ParseUint(d.Id(), 10, 0)
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
	config := meta.(*Configuration)
	controller := config.Controller
	var gid int

	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		group := d.Get("group").(string)
		gid, err = controller.Groups().ByName(group)
		if err != nil {
			return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
		}
	} else {
		gid = d.Get("gid").(int)
	}

	err = vmc.Chown(-1, gid)
	if err != nil {
		return fmt.Errorf("Can't find a group with ID `%d`: %s", gid, err)
	}

	return nil
}

func resourceOpennebulaVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

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
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve informations",
				Detail:   err.Error(),
			})
			return diags
		}

		tplContext, _ := tpl.Template.GetVector(vmk.ContextVec)

		// customize template except for memory and cpu.
		vmTpl, err := generateVm(d, tplContext)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate description",
				Detail:   err.Error(),
			})
			return diags
		}

		generateVMNIC(d, vmTpl)

		// Instantiate template without creating a persistent copy of the template
		// Note that the new VM is not pending
		vmID, err = tc.Instantiate(d.Get("name").(string), d.Get("pending").(bool), vmTpl.String(), false)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to instantiate the template",
				Detail:   err.Error(),
			})
			return diags
		}
	} else {
		if _, ok := d.GetOk("cpu"); !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "cpu is not defined",
				Detail:   "cpu is mandatory when template_id is not defined",
			})
		}
		if _, ok := d.GetOk("memory"); !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "memory is not defined",
				Detail:   "memory is mandatory when template_id is not defined",
			})
		}
		if len(diags) > 0 {
			return diags
		}

		vmTpl, err := generateVm(d, nil)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate description",
				Detail:   err.Error(),
			})
			return diags
		}

		generateVMNIC(d, vmTpl)

		// Create VM not in pending state
		vmID, err = controller.VMs().Create(vmTpl.String(), d.Get("pending").(bool))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	d.SetId(fmt.Sprintf("%v", vmID))
	vmc := controller.VM(vmID)

	expectedState := "RUNNING"
	if d.Get("pending").(bool) {
		expectedState = "HOLD"
	}

	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultVMTimeout {
		timeout = d.Timeout(schema.TimeoutCreate)
	}
	_, err = waitForVMState(ctx, vmc, timeout, expectedState)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to wait virtual machine to be in %s state", expectedState),
			Detail:   err.Error(),
		})
		return diags
	}

	//Set the permissions on the VM if it was defined, otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = vmc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",

				Detail: err.Error(),
			})
			return diags
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err = changeVmGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",

				Detail: err.Error(),
			})
			return diags
		}
	}

	if lock, ok := d.GetOk("lock"); ok && lock.(string) != "UNLOCK" {

		var level shared.LockLevel
		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to convert lock level",
				Detail:   err.Error(),
			})
			return diags
		}

		err = vmc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   err.Error(),
			})
			return diags
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

		return resourceOpennebulaVirtualMachineReadCustom(ctx, d, meta, func(ctx context.Context, d *schema.ResourceData, vmInfos *vm.VM) diag.Diagnostics {

			err := flattenDiskFunc(d, &vmInfos.Template)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to flatten disks",
					Detail:   err.Error(),
				})
				return diags
			}

			err = flattenNICFunc(d, &vmInfos.Template)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to flatten NICs",
					Detail:   err.Error(),
				})
				return diags
			}
			return nil
		})

	}

	d.Set("template_nic", []interface{}{})
	d.Set("template_disk", []interface{}{})

	return resourceOpennebulaVirtualMachineRead(ctx, d, meta)
}

func resourceOpennebulaVirtualMachineReadCustom(ctx context.Context, d *schema.ResourceData, meta interface{}, customVM customVMFunc) diag.Diagnostics {

	var diags diag.Diagnostics

	vmc, err := getVirtualMachineController(d, meta, -2, -1, -1)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual machine %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get controller",
			Detail:   err.Error(),
		})
		return diags

	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vm, err := vmc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   err.Error(),
		})
		return diags
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
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set attribute",
			Detail:   err.Error(),
		})
		return diags
	}

	if customVM != nil {
		customDiags := customVM(ctx, d, vm)
		if len(customDiags) > 0 {
			return customDiags
		}
	}

	err = flattenTemplate(d, &vm.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten",
			Detail:   err.Error(),
		})
		return diags
	}

	err = flattenVMUserTemplate(d, &vm.UserTemplate.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten template",
			Detail:   err.Error(),
		})
		return diags
	}

	err = flattenTags(d, &vm.UserTemplate)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten tags",
			Detail:   err.Error(),
		})
		return diags
	}

	if vm.LockInfos != nil {
		d.Set("lock", LockLevelToString(vm.LockInfos.Locked))
	}

	return nil
}

func resourceOpennebulaVirtualMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceOpennebulaVirtualMachineReadCustom(ctx, d, meta, func(ctx context.Context, d *schema.ResourceData, vmInfos *vm.VM) diag.Diagnostics {

		var diags diag.Diagnostics

		err := flattenVMDisk(d, &vmInfos.Template)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to flatten disks",
				Detail:   err.Error(),
			})
			return diags
		}
		err = flattenVMNIC(d, &vmInfos.Template)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to flatten NICs",
				Detail:   err.Error(),
			})
			return diags
		}

		return nil
	})
}

func flattenVMDiskComputed(diskConfig map[string]interface{}, disk shared.Disk) map[string]interface{} {

	diskMap := flattenDiskComputed(disk)

	if diskConfig["size"].(int) > 0 {
		diskMap["size"] = diskMap["computed_size"]
	}
	if len(diskConfig["target"].(string)) > 0 {
		diskMap["target"] = diskMap["computed_target"]
	}
	if len(diskConfig["driver"].(string)) > 0 {
		diskMap["driver"] = diskMap["computed_driver"]
	}
	if len(diskConfig["volatile_format"].(string)) > 0 {
		diskMap["volatile_format"] = diskMap["computed_volatile_format"]
	}

	return diskMap
}

func flattenDiskComputed(disk shared.Disk) map[string]interface{} {
	size, _ := disk.GetI(shared.Size)
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	volatileFormat, _ := disk.Get("FORMAT")
	diskID, _ := disk.GetI(shared.DiskID)

	return map[string]interface{}{
		"disk_id":                  diskID,
		"computed_size":            size,
		"computed_target":          target,
		"computed_driver":          driver,
		"computed_volatile_format": volatileFormat,
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

func matchDisk(diskConfig map[string]interface{}, disk shared.Disk) bool {

	size, _ := disk.GetI(shared.Size)
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	volatileType, _ := disk.Get("TYPE")
	volatileFormat, _ := disk.Get("FORMAT")

	return emptyOrEqual(diskConfig["target"], target) &&
		emptyOrEqual(diskConfig["size"], size) &&
		emptyOrEqual(diskConfig["driver"], driver) &&
		emptyOrEqual(diskConfig["volatile_type"], volatileType) &&
		emptyOrEqual(diskConfig["volatile_format"], volatileFormat)
}

func matchDiskComputed(diskConfig map[string]interface{}, disk shared.Disk) bool {

	size, _ := disk.GetI(shared.Size)
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	format, _ := disk.Get("FORMAT")

	return (target == diskConfig["computed_target"].(string)) &&
		(size == diskConfig["computed_size"].(int)) &&
		(driver == diskConfig["computed_driver"].(string)) &&
		(format == diskConfig["computed_volatile_format"].(string))

}

// flattenVMDisk is similar to flattenDisk but deal with computed_* attributes
// this is a temporary solution until we can use nested attributes marked computed and optional
func flattenVMDisk(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set disks to resource
	disks := vmTemplate.GetDisks()
	diskConfigs := d.Get("disk").([]interface{})

	diskList := make([]interface{}, 0, len(disks))

diskLoop:
	for _, disk := range disks {

		// exclude disk from template_disk
		tplDiskConfigs := d.Get("template_disk").([]interface{})
		for _, tplDiskConfigIf := range tplDiskConfigs {
			tplDiskConfig := tplDiskConfigIf.(map[string]interface{})

			if matchDiskComputed(tplDiskConfig, disk) {
				continue diskLoop
			}
		}

		// copy disk config values
		var diskMap map[string]interface{}

		match := false
		for j := 0; j < len(diskConfigs); j++ {
			diskConfig := diskConfigs[j].(map[string]interface{})

			// try to reidentify the disk based on it's configuration values
			if !matchDisk(diskConfig, disk) {
				continue
			}

			match = true
			diskMap = flattenVMDiskComputed(diskConfig, disk)

			imageID, _ := disk.GetI(shared.ImageID)
			diskMap["image_id"] = imageID

			// for volatile disk, TYPE has the same value
			// than DISK_TYPE
			if imageID == -1 {
				volatileType, _ := disk.Get("TYPE")
				diskMap["volatile_type"] = volatileType
			}

			diskList = append(diskList, diskMap)

			break

		}

		if !match {
			ID, _ := disk.ID()
			log.Printf("[WARN] Configuration for disk ID: %d not found.", ID)
		}

	}

	if len(diskList) > 0 {
		err := d.Set("disk", diskList)
		if err != nil {
			return err
		}
	}

	return nil
}

func flattenNICComputed(nic shared.NIC) map[string]interface{} {
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

func flattenVMNICComputed(NICConfig map[string]interface{}, NIC shared.NIC) map[string]interface{} {

	NICMap := flattenNICComputed(NIC)

	if len(NICConfig["ip"].(string)) > 0 {
		NICMap["ip"] = NICMap["computed_ip"]
	}
	if len(NICConfig["mac"].(string)) > 0 {
		NICMap["mac"] = NICMap["computed_mac"]
	}
	if len(NICConfig["model"].(string)) > 0 {
		NICMap["model"] = NICMap["computed_model"]
	}
	if len(NICConfig["virtio_queues"].(string)) > 0 {
		NICMap["virtio_queues"] = NICMap["computed_virtio_queues"]
	}
	if len(NICConfig["physical_device"].(string)) > 0 {
		NICMap["physical_device"] = NICMap["computed_physical_device"]
	}
	if len(NICConfig["security_groups"].([]interface{})) > 0 {
		NICMap["security_groups"] = NICMap["computed_security_groups"]
	}

	return NICMap
}

// flattenVMTemplateNIC read NIC that come from template when instantiating a VM
func flattenVMTemplateNIC(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set Nics to resource
	nics := vmTemplate.GetNICs()
	nicList := make([]interface{}, 0, len(nics))

	for i, nic := range nics {

		networkID, _ := nic.GetI(shared.NetworkID)
		nicRead := flattenNICComputed(nic)
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

func matchNIC(NICConfig map[string]interface{}, NIC shared.NIC) bool {

	ip, _ := NIC.Get(shared.IP)
	mac, _ := NIC.Get(shared.MAC)
	physicalDevice, _ := NIC.GetStr("PHYDEV")

	model, _ := NIC.Get(shared.Model)
	virtioQueues, _ := NIC.GetStr("VIRTIO_QUEUES")
	securityGroupsArray, _ := NIC.Get(shared.SecurityGroups)

	if NICConfig["security_groups"] != nil && len(NICConfig["security_groups"].([]interface{})) > 0 {

		sg := strings.Split(securityGroupsArray, ",")
		sgConfig := NICConfig["security_groups"].([]interface{})

		if len(sg) != len(sgConfig) {
			return false
		}

		for i := 0; i < len(sg); i++ {
			sgInt, err := strconv.ParseInt(sg[i], 10, 0)
			if err != nil {
				return false
			}
			if int(sgInt) != sgConfig[i].(int) {
				return false
			}
		}

	}

	return emptyOrEqual(NICConfig["ip"], ip) &&
		emptyOrEqual(NICConfig["mac"], mac) &&
		emptyOrEqual(NICConfig["physical_device"], physicalDevice) &&
		emptyOrEqual(NICConfig["model"], model) &&
		emptyOrEqual(NICConfig["virtio_queues"], virtioQueues)
}

func matchNICComputed(NICConfig map[string]interface{}, NIC shared.NIC) bool {
	ip, _ := NIC.Get(shared.IP)
	mac, _ := NIC.Get(shared.MAC)
	physicalDevice, _ := NIC.GetStr("PHYDEV")

	model, _ := NIC.Get(shared.Model)
	virtioQueues, _ := NIC.GetStr("VIRTIO_QUEUES")
	securityGroupsArray, _ := NIC.Get(shared.SecurityGroups)

	sg := strings.Split(securityGroupsArray, ",")
	sgConfig := NICConfig["computed_security_groups"].([]interface{})

	if len(sg) != len(sgConfig) {
		return false
	}

	for i := 0; i < len(sg); i++ {
		sgInt, err := strconv.ParseInt(sg[i], 10, 0)
		if err != nil {
			return false
		}
		if int(sgInt) != sgConfig[i].(int) {
			return false
		}
	}

	return ip == NICConfig["computed_ip"].(string) &&
		mac == NICConfig["computed_mac"].(string) &&
		physicalDevice == NICConfig["computed_physical_device"].(string) &&
		model == NICConfig["computed_model"].(string) &&
		virtioQueues == NICConfig["computed_virtio_queues"].(string)
}

// flattenVMNIC is similar to flattenNIC but deal with computed_* attributes
// this is a temporary solution until we can use nested attributes marked computed and optional
func flattenVMNIC(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set Nics to resource
	nics := vmTemplate.GetNICs()
	nicsConfigs := d.Get("nic").([]interface{})

	nicList := make([]interface{}, 0, len(nics))

NICLoop:
	for i, nic := range nics {

		// exclude NIC listed in template_nic
		tplNICConfigs := d.Get("template_nic").([]interface{})
		for _, tplNICConfigIf := range tplNICConfigs {
			tplNICConfig := tplNICConfigIf.(map[string]interface{})

			if matchNICComputed(tplNICConfig, nic) {
				continue NICLoop
			}
		}

		// copy nic config values
		var nicMap map[string]interface{}

		match := false
		for j := 0; j < len(nicsConfigs); j++ {
			nicConfig := nicsConfigs[j].(map[string]interface{})

			// try to reidentify the nic based on it's configuration values
			// network_id is not sufficient in case of a network attached twice
			if !matchNIC(nicConfig, nic) {
				continue
			}

			match = true
			nicMap = flattenVMNICComputed(nicConfig, nic)

			networkID, _ := nic.GetI(shared.NetworkID)
			nicMap["network_id"] = networkID

			nicList = append(nicList, nicMap)

			break

		}

		if !match {
			ID, _ := nic.ID()
			log.Printf("[WARN] Configuration for NIC ID: %d not found.", ID)
		}

		if i == 0 {
			d.Set("ip", nicMap["computed_ip"])
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

	tagsInterface := d.Get("tags").(map[string]interface{})

	tags := pairsToMapFilter(vmUserTpl.Template, tagsInterface)
	if len(tags) > 0 {
		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceOpennebulaVirtualMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	serviceTemplateID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.VM(int(serviceTemplateID)).Info(false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	err := resourceOpennebulaVirtualMachineUpdateCustom(ctx, d, meta, customVirtualMachineUpdate)
	if err != nil {
		return err
	}

	return resourceOpennebulaVirtualMachineRead(ctx, d, meta)
}

func customVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	if d.HasChange("nic") {
		err := updateNIC(ctx, d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update NIC",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	return nil
}

func resourceOpennebulaVirtualMachineUpdateCustom(ctx context.Context, d *schema.ResourceData, meta interface{}, customFunc customFunc) diag.Diagnostics {

	var diags diag.Diagnostics

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vmInfos, err := vmc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   err.Error(),
		})
		return diags
	}

	lock, lockOk := d.GetOk("lock")
	if d.HasChange("lock") && lockOk && lock.(string) == "UNLOCK" {

		err = vmc.Unlock()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to unlock",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	if d.HasChange("name") {
		err := vmc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   err.Error(),
			})
			return diags
		}
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		vmInfos, err := vmc.Info(false)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve informations",
				Detail:   err.Error(),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name (%s) for VM ID %x\n", vmInfos.Name, vmInfos.ID)
	}

	if d.HasChange("permissions") && d.Get("permissions") != "" {
		if perms, ok := d.GetOk("permissions"); ok {
			err = vmc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change permissions",
					Detail:   err.Error(),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated Permissions VM %s\n", vmInfos.Name)
	}

	if d.HasChange("group") {
		err := changeVmGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change VM group",
				Detail:   err.Error(),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated group for VM %s\n", vmInfos.Name)
	}

	update := false
	tpl := &vm.Template{
		Template: dyn.Template{
			Elements: vmInfos.UserTemplate.Template.Elements,
		},
	}

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

	if d.HasChange("tags") {

		oldTagsIf, newTagsIf := d.GetChange("tags")
		oldTags := oldTagsIf.(map[string]interface{})
		newTags := newTagsIf.(map[string]interface{})

		// delete tags
		for k, _ := range oldTags {
			_, ok := newTags[k]
			if ok {
				continue
			}
			tpl.Del(strings.ToUpper(k))
		}

		// add/update tags
		for k, v := range newTags {
			tpl.Template.Del(strings.ToUpper(k))
			tpl.AddPair(strings.ToUpper(k), v)
		}

		update = true
	}

	if update {
		err = vmc.Update(tpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	if d.HasChange("disk") {
		err = updateDisk(ctx, d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update disk",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	if customFunc != nil {
		customDiags := customFunc(ctx, d, meta)
		if len(customDiags) > 0 {
			return customDiags
		}
	}

	updateConf := false

	// retrieve only template sections managed by updateconf method
	tpl = vm.NewTemplate()
	for _, name := range []string{"OS", "FEATURES", "INPUT", "GRAPHICS", "RAW", "CONTEXT"} {
		vectors := vmInfos.Template.GetVectors(name)
		for _, vec := range vectors {
			tpl.Elements = append(tpl.Elements, vec)
		}
	}

	if d.HasChange("os") {
		updateConf = true

		log.Printf("[DEBUG] Update os")

		old, new := d.GetChange("os")
		newOSSlice := new.([]interface{})

		if len(newOSSlice) == 0 {
			// No os configuration to apply
			tpl.Del("OS")
		} else {

			appliedOSSlice := old.([]interface{})

			if len(appliedOSSlice) == 0 {
				// No os configuration applied
				addOS(tpl, newOSSlice)
			} else {

				newOS := newOSSlice[0].(map[string]interface{})
				appliedOS := appliedOSSlice[0].(map[string]interface{})

				err := updateVMTemplateVec(tpl, "OS", appliedOS, newOS)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to update OS vector",
						Detail:   err.Error(),
					})
					return diags
				}
			}
		}
	}

	if d.HasChange("graphics") {
		updateConf = true

		log.Printf("[DEBUG] Update graphics")

		old, new := d.GetChange("graphics")
		newGraphicsSlice := new.([]interface{})

		if len(newGraphicsSlice) == 0 {
			// No graphics configuration to apply
			tpl.Del("GRAPHICS")
		} else {

			appliedGraphicsSlice := old.([]interface{})

			if len(appliedGraphicsSlice) == 0 {
				// No graphics configuration applied
				addGraphic(tpl, newGraphicsSlice)
			} else {

				newGraphics := newGraphicsSlice[0].(map[string]interface{})
				appliedGraphics := appliedGraphicsSlice[0].(map[string]interface{})

				updateVMTemplateVec(tpl, "GRAPHICS", appliedGraphics, newGraphics)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to update GRAPHICS vector",
						Detail:   err.Error(),
					})
					return diags
				}
			}
		}
	}

	if d.HasChange("context") {

		updateConf = true

		log.Printf("[DEBUG] Update context")

		old, new := d.GetChange("context")
		appliedContext := old.(map[string]interface{})
		newContext := new.(map[string]interface{})

		if len(newContext) == 0 {
			// No context configuration to apply
			tpl.Del(vmk.ContextVec)
		} else {

			var contextVec *dyn.Vector
			if len(appliedContext) == 0 {
				// No context configuration applied
				contextVec = tpl.AddVector(vmk.ContextVec)

				// Add new elements
				for key, value := range newContext {
					keyUp := strings.ToUpper(key)

					_, ok := appliedContext[keyUp]
					if ok {
						continue
					}

					contextVec.AddPair(keyUp, value)
				}

			} else {
				updateVMTemplateVec(tpl, "CONTEXT", appliedContext, newContext)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to update CONTEXT vector",
						Detail:   err.Error(),
					})
					return diags
				}
			}
		}
	}

	if d.HasChange("cpu") || d.HasChange("vcpu") || d.HasChange("memory") {

		timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
		if timeout == defaultVMTimeout {
			timeout = d.Timeout(schema.TimeoutUpdate)
		}

		vmState, _, _ := vmInfos.State()
		vmRequireShutdown := vmState != vm.Poweroff && vmState != vm.Undeployed
		if vmRequireShutdown {
			if d.Get("hard_shutdown").(bool) {
				err = vmc.PoweroffHard()
			} else {
				err = vmc.Poweroff()
			}
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to power off",
					Detail:   err.Error(),
				})
				return diags
			}
			_, err = waitForVMState(ctx, vmc, timeout, "POWEROFF")
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to wait virtual machine to be in POWEROFF state",
					Detail:   err.Error(),
				})
				return diags
			}
		}

		resizeTpl := dyn.NewTemplate()
		cpu := d.Get("cpu").(float64)
		if cpu > 0 {
			resizeTpl.AddPair("CPU", cpu)
		}

		vcpu := d.Get("vcpu").(int)
		if vcpu > 0 {
			resizeTpl.AddPair("VCPU", vcpu)
		}

		memory := d.Get("memory").(int)
		if cpu > 0 {
			resizeTpl.AddPair("MEMORY", memory)
		}

		err = vmc.Resize(resizeTpl.String(), true)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to resize",
				Detail:   err.Error(),
			})
			return diags
		}

		if vmRequireShutdown {
			err = vmc.Resume()
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to resume",
					Detail:   err.Error(),
				})
				return diags
			}
			_, err = waitForVMState(ctx, vmc, timeout, "RUNNING")
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to wait virtual machine to be in RUNNING state",
					Detail:   err.Error(),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully resized VM %s\n", vmInfos.Name)
	}

	if updateConf {

		timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
		if timeout == defaultVMTimeout {
			timeout = d.Timeout(schema.TimeoutUpdate)
		}

		_, err = waitForVMState(ctx, vmc, timeout, "RUNNING")
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to wait virtual machine to be in RUNNING state",
				Detail:   err.Error(),
			})
			return diags
		}

		log.Printf("[INFO] Update VM configuration: %s", tpl.String())

		err := vmc.UpdateConf(tpl.String())
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   err.Error(),
			})
			return diags
		}

		_, err = waitForVMState(ctx, vmc, timeout, "RUNNING")
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to wait virtual machine to be in RUNNING state",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	if d.HasChange("lock") && lockOk && lock.(string) != "UNLOCK" {

		var level shared.LockLevel

		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to convert lock level",
				Detail:   err.Error(),
			})
			return diags
		}
		err = vmc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	return resourceOpennebulaVirtualMachineRead(ctx, d, meta)
}

func updateDisk(ctx context.Context, d *schema.ResourceData, meta interface{}) error {

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Update disk configuration")

	old, new := d.GetChange("disk")
	attachedDisksCfg := old.([]interface{})
	newDisksCfg := new.([]interface{})

	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultVMTimeout {
		timeout = d.Timeout(schema.TimeoutUpdate)
	}

	// get unique elements of each list of configs
	// NOTE: diffListConfig relies on Set, so we may loose list ordering of disks here
	// it's why we reorder the attach list below
	toDetach, toAttach := diffListConfig(newDisksCfg, attachedDisksCfg,
		&schema.Resource{
			Schema: diskFields(),
		},
		"image_id",
		"target",
		"driver",
		"volatile_type",
		"volatile_format")

	// reorder toAttach disk list according to new disks list order
	newDisktoAttach := make([]interface{}, len(toAttach))
	i := 0
	for _, newDiskIf := range newDisksCfg {
		newDisk := newDiskIf.(map[string]interface{})

		for _, diskToAttachIf := range toAttach {
			disk := diskToAttachIf.(map[string]interface{})

			// if disk have the same attributes
			if (disk["target"] == newDisk["target"]) &&
				disk["size"] == newDisk["size"] &&
				disk["driver"] == newDisk["driver"] {

				newDisktoAttach[i] = disk
				i++
				break
			}
		}
	}
	toAttach = newDisktoAttach

	// get disks to resize
	_, toResize := diffListConfig(newDisksCfg, attachedDisksCfg,
		&schema.Resource{
			Schema: diskFields(),
		},
		"image_id",
		"size")

	// Detach the disks
	for _, diskIf := range toDetach {
		diskConfig := diskIf.(map[string]interface{})

		// ignore disk without image_id and type
		if diskConfig["image_id"].(int) == -1 &&
			len(diskConfig["volatile_type"].(string)) == 0 {

			log.Printf("[INFO] ignore disk without image_id and type")
			continue
		}

		diskID := diskConfig["disk_id"].(int)

		err := vmDiskDetach(ctx, vmc, timeout, diskID)
		if err != nil {
			return fmt.Errorf("vm disk detach: %s", err)

		}
	}

	// Attach the disks
	for _, diskIf := range toAttach {
		diskConfig := diskIf.(map[string]interface{})

		// ignore disk without image_id and type
		if diskConfig["image_id"].(int) == -1 &&
			len(diskConfig["volatile_type"].(string)) == 0 {

			log.Printf("[INFO] ignore disk without image_id and type")
			continue
		}

		diskTpl := makeDiskVector(diskConfig)

		_, err := vmDiskAttach(ctx, vmc, timeout, diskTpl)
		if err != nil {
			return fmt.Errorf("vm disk attach: %s", err)
		}
	}

	// Resize disks
	for _, diskIf := range toResize {
		diskConfig := diskIf.(map[string]interface{})

		// ignore disk without image_id and type
		if diskConfig["image_id"].(int) == -1 &&
			len(diskConfig["volatile_type"].(string)) == 0 {

			log.Printf("[INFO] ignore disk without image_id and type")
			continue
		}

		// retrieve the the disk_id
		for _, d := range attachedDisksCfg {

			cfg := d.(map[string]interface{})
			if diskConfig["image_id"].(int) != cfg["image_id"].(int) ||
				(len(diskConfig["target"].(string)) > 0 && diskConfig["target"] != cfg["computed_target"]) ||
				(len(diskConfig["driver"].(string)) > 0 && diskConfig["driver"] != cfg["computed_driver"]) ||
				diskConfig["size"].(int) <= cfg["computed_size"].(int) {

				continue
			}

			diskID := cfg["disk_id"].(int)

			err := vmDiskResize(ctx, vmc, timeout, diskID, diskConfig["size"].(int))
			if err != nil {
				return fmt.Errorf("vm disk resize: %s", err)
			}

		}
	}

	return nil
}

func updateNIC(ctx context.Context, d *schema.ResourceData, meta interface{}) error {

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Update NIC configuration")

	old, new := d.GetChange("nic")
	attachedNicsCfg := old.([]interface{})
	newNicsCfg := new.([]interface{})

	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultVMTimeout {
		timeout = d.Timeout(schema.TimeoutUpdate)
	}

	// get unique elements of each list of configs
	// NOTE: diffListConfig relies on Set, so we may loose list ordering of NICs here
	// it's why we reorder the attach list below
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

	// in case of NICs updated in the middle of the NIC list
	// they would be reattached at the end of the list (we don't have in place XML-RPC update method).
	// keep_nic_order prevent this behavior adding more NICs to detach/attach to keep initial ordering
	if d.Get("keep_nic_order").(bool) && len(toDetach) > 0 {

		// retrieve the minimal nic ID to detach
		firstNIC := toDetach[0].(map[string]interface{})
		minID := firstNIC["nic_id"].(int)
		for _, nicIf := range toDetach[1:] {
			nicConfig := nicIf.(map[string]interface{})

			nicID := nicConfig["nic_id"].(int)
			if nicID < minID {
				minID = nicID
			}
		}

		// NICs with greater nic ID should be detached
	oldNICLoop:
		for _, nicIf := range attachedNicsCfg {
			nicConfig := nicIf.(map[string]interface{})

			// collect greater nic IDs
			nicID := nicConfig["nic_id"].(int)
			if nicID > minID {

				// add the nic if not already present in toDetach
				for _, nicDetachIf := range toDetach {
					nicDetachConfig := nicDetachIf.(map[string]interface{})

					// nic is already present
					detachNICID := nicDetachConfig["nic_id"].(int)
					if detachNICID == nicID {
						continue oldNICLoop
					}
				}

				// add the NIC to detach it
				toDetach = append(toDetach, nicConfig)

				// add the NIC to reattach it
				toAttach = append(toAttach, nicConfig)
			}
		}

	}

	// reorder toAttach NIC list according to new nics list order
	newNICtoAttach := make([]interface{}, len(toAttach))
	i := 0
	for _, newNICIf := range newNicsCfg {
		newNIC := newNICIf.(map[string]interface{})
		newNICSecGroup := newNIC["security_groups"].([]interface{})

		for _, NICToAttachIf := range toAttach {
			NIC := NICToAttachIf.(map[string]interface{})

			// if NIC have the same attributes

			// compare security_groups
			NICSecGroup := NIC["security_groups"].([]interface{})

			if ArrayToString(NICSecGroup, ",") != ArrayToString(newNICSecGroup, ",") {
				continue
			}

			// compare other attributes
			if (NIC["ip"] == newNIC["ip"] &&
				NIC["mac"] == newNIC["mac"]) &&
				NIC["model"] == newNIC["model"] &&
				NIC["virtio_queues"] == newNIC["virtio_queues"] &&
				NIC["physical_device"] == newNIC["physical_device"] {

				newNICtoAttach[i] = NIC
				i++
				break
			}
		}
	}
	toAttach = newNICtoAttach

	// Detach the nics
	for _, nicIf := range toDetach {
		nicConfig := nicIf.(map[string]interface{})

		nicID := nicConfig["nic_id"].(int)

		err := vmNICDetach(ctx, vmc, timeout, nicID)
		if err != nil {
			return fmt.Errorf("vm nic detach: %s", err)

		}
	}

	// Attach the nics
	for _, nicIf := range toAttach {

		nicConfig := nicIf.(map[string]interface{})

		nicTpl := makeNICVector(nicConfig)

		_, err := vmNICAttach(ctx, vmc, timeout, nicTpl)
		if err != nil {
			return fmt.Errorf("vm nic attach: %s", err)
		}
	}

	return nil
}

// updateVMVec update a vector of an existing VM template
func updateVMTemplateVec(tpl *vm.Template, vecName string, appliedCfg, newCfg map[string]interface{}) error {

	// Retrieve vector
	var targetVec *dyn.Vector
	vectors := tpl.GetVectors(vecName)
	switch len(vectors) {
	case 0:
		return fmt.Errorf("No %s vector present", vecName)
	case 1:
		targetVec = vectors[0]

		// Remove or update existing elements
		for key := range appliedCfg {
			keyUp := strings.ToUpper(key)

			value, ok := newCfg[keyUp]
			if ok {
				// update existing element
				targetVec.Del(keyUp)
				targetVec.AddPair(keyUp, fmt.Sprint(value))
			} else {
				// remove element
				targetVec.Del(keyUp)
			}
		}
	default:
		return fmt.Errorf("Multiple %s vectors", vecName)
	}

	// Add new elements
	for key, value := range newCfg {
		keyUp := strings.ToUpper(key)

		_, ok := appliedCfg[keyUp]
		if ok {
			continue
		}

		targetVec.AddPair(keyUp, value)
	}

	return nil
}

func resourceOpennebulaVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// wait state to be ready
	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultVMTimeout {
		timeout = d.Timeout(schema.TimeoutUpdate)
	}

	_, err = waitForVMState(ctx, vmc, timeout, vmDeleteReadyStates...)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to wait virtual machine to be in %s state", strings.Join(vmDeleteReadyStates, " ")),
			Detail:   err.Error(),
		})
		return diags
	}

	if d.Get("hard_shutdown").(bool) {
		err = vmc.TerminateHard()
	} else {
		err = vmc.Terminate()
	}
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to terminate",
			Detail:   err.Error(),
		})
		return diags
	}

	ret, err := waitForVMState(ctx, vmc, timeout, "DONE")
	if err != nil {

		log.Printf("[WARN] %s\n", err)

		// Retry if timeout not reached
		_, ok := err.(*resource.TimeoutError)
		if !ok && ret != nil {

			vmInfos, _ := ret.(*vm.VM)
			vmState, vmLcmState, _ := vmInfos.State()
			if vmState == vm.Active && vmLcmState == vm.EpilogFailure {

				log.Printf("[INFO] retry terminate VM\n")

				if d.Get("hard_shutdown").(bool) {
					err = vmc.TerminateHard()
				} else {
					err = vmc.Terminate()
				}
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to terminate",
						Detail:   err.Error(),
					})
					return diags
				}

				_, err = waitForVMState(ctx, vmc, timeout, "DONE")
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to wait virtual machine to be in DONE state",
						Detail:   err.Error(),
					})
					return diags
				}

			} else {
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to wait virtual machine to be in DONE state",
						Detail:   err.Error(),
					})
					return diags
				}
			}
		}
	}

	log.Printf("[INFO] Successfully terminated VM\n")
	return nil
}

func waitForVMState(ctx context.Context, vmc *goca.VMController, timeout time.Duration, states ...string) (interface{}, error) {

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

			case vm.Done, vm.Hold, vm.Suspended, vm.Stopped, vm.Poweroff, vm.Undeployed:
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
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForStateContext(ctx)

}

func waitForVMsStates(ctx context.Context, c *goca.Controller, vmIDs []int, timeout time.Duration, states ...string) ([]interface{}, []error) {

	errors := make([]error, 0)
	vmsInfos := make([]interface{}, 0)

	for _, id := range vmIDs {
		vmInfo, err := waitForVMState(ctx, c.VM(id), timeout, states...)
		if vmInfo != nil {
			vmsInfos = append(vmsInfos, vmInfo)
		}
		if err != nil {
			errors = append(errors, err)
		}
	}

	return vmsInfos, errors
}

func generateVm(d *schema.ResourceData, tplContext *dyn.Vector) (*vm.Template, error) {

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

	err := generateVMTemplate(d, tpl)
	if err != nil {
		return tpl, err
	}

	return tpl, nil
}

func generateVMNIC(d *schema.ResourceData, tpl *vm.Template) {
	//Generate NIC definition
	nics := d.Get("nic").([]interface{})
	log.Printf("Number of NICs: %d", len(nics))

	for i := 0; i < len(nics); i++ {
		nicconfig := nics[i].(map[string]interface{})

		nic := makeNICVector(nicconfig)
		tpl.Elements = append(tpl.Elements, nic)
	}
}

func resourceVMCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {

	onChange := diff.Get("on_disk_change").(string)

	if strings.ToUpper(onChange) == "RECREATE" {

		oldDisk, newDisk := diff.GetChange("disk")
		oldDiskList := oldDisk.([]interface{})
		newDiskList := newDisk.([]interface{})
		toDetach, _ := diffListConfig(newDiskList, oldDiskList,
			&schema.Resource{
				Schema: diskFields(),
			},
			"image_id",
			"target",
			"driver")

		if len(toDetach) > 0 {
			for i := range oldDiskList {
				diff.ForceNew(fmt.Sprintf("disk.%d.image_id", i))
				diff.ForceNew(fmt.Sprintf("disk.%d.target", i))
				diff.ForceNew(fmt.Sprintf("disk.%d.driver", i))
			}
		}
	}

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
