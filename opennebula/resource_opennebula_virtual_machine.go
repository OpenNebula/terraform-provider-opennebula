package opennebula

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	errs "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

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
				ForceNew:    true,
				Description: "Id of the VM template to use",
			},
			"pending": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Pending state of the VM during its creation, by default it is set to false",
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
			"cpu": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Computed:    true,
				Description: "Amount of CPU quota assigned to the virtual machine",
			},
			"vcpu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Number of virtual CPUs assigned to the virtual machine",
			},
			"memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Amount of memory (RAM) in MB assigned to the virtual machine",
			},
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Description: "Context variables",
			},
			"disk": {
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
			},
			"graphics": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				MinItems:    0,
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
			},
			"nic": {
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
			},
			"os": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				//Computed:    true,
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
			},
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
			"vmgroup": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				MinItems:    0,
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
			},
		},
	}
}

func getVirtualMachineController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.VMController, error) {
	controller := meta.(*goca.Controller)
	var vmc *goca.VMController

	// Try to find the VM by ID, if specified
	if d.Id() != "" {
		id, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		vmc = controller.VM(int(id))
	}

	// Otherwise, try to find the VM by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.VMs().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
		}
		vmc = controller.VM(gid)
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

	if v, ok := d.GetOk("template_id"); ok {
		// if template id is set, instantiate a VM from this template
		tc := controller.Template(v.(int))

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
	}

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", vmID))
	vmc := controller.VM(vmID)

	expectedState := "running"
	if d.Get("pending").(bool) {
		expectedState = "hold"
	}

	_, err = waitForVmState(d, meta, expectedState)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for virtual machine (%s) to be in state %s: %s", expectedState, d.Id(), err)
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

	return resourceOpennebulaVirtualMachineRead(d, meta)
}

func resourceOpennebulaVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	vmc, err := getVirtualMachineController(d, meta, -2, -1, -1)
	if err != nil {
		switch err.(type) {
		case *errs.ClientError:
			clientErr, _ := err.(*errs.ClientError)
			if clientErr.Code == errs.ClientRespHTTP {
				response := clientErr.GetHTTPResponse()
				if response.StatusCode == http.StatusNotFound {
					log.Printf("[WARN] Removing virtual machine %s from state because it no longer exists in", d.Get("name"))
					d.SetId("")
					return nil
				}
			}
			return err
		default:
			return err
		}
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

	err = flattenTemplate(d, &vm.Template)
	if err != nil {
		return err
	}
	return nil
}

func flattenTemplate(d *schema.ResourceData, vmTemplate *vm.Template) error {

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

	// Set OVM Group to resource
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

	// Enable partial state mode
	d.Partial(true)

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
		d.SetPartial("name")
		log.Printf("[INFO] Successfully updated name (%s) for VM ID %x\n", vm.Name, vm.ID)
	}

	if d.HasChange("permissions") && d.Get("permissions") != "" {
		if perms, ok := d.GetOk("permissions"); ok {
			err = vmc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		d.SetPartial("permissions")
		log.Printf("[INFO] Successfully updated Permissions VM %s\n", vm.Name)
	}

	if d.HasChange("group") {
		err := changeVmGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for VM %s\n", vm.Name)
	}

	// We succeeded, disable partial mode. This causes Terraform to save
	// save all fields again.
	d.Partial(false)

	return nil
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

	_, err = waitForVmState(d, meta, "done")
	if err != nil {
		vm, _ := vmc.Info(false)

		vmState, vmLcmState, _ := vm.State()
		if vmLcmState.String() == "EPILOG_FAILURE" {
			if err = vmc.TerminateHard(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf(
				"Error waiting for virtual machine (%s) to be in state DONE: %s (state: %v, lcmState: %v)", d.Id(), err, vmState, vmLcmState)
		}
	}

	log.Printf("[INFO] Successfully terminated VM\n")
	return nil
}

func waitForVmState(d *schema.ResourceData, meta interface{}, state string) (interface{}, error) {
	var vm *vm.VM
	var err error
	//Get VM controller
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return vm, err
	}

	log.Printf("Waiting for VM (%s) to be in state Done", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{"anythingelse"}, Target: []string{state},
		Refresh: func() (interface{}, string, error) {
			log.Println("Refreshing VM state...")
			if d.Id() != "" {
				//Get VM controller
				vmc, err = getVirtualMachineController(d, meta)
				if err != nil {
					return vm, "", fmt.Errorf("Could not find VM by ID %s", d.Id())
				}
			}
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			vm, err = vmc.Info(false)
			if err != nil {
				if strings.Contains(err.Error(), "Error getting") {
					return vm, "notfound", nil
				}
				return vm, "", err
			}
			vmState, vmLcmState, err := vm.State()
			if err != nil {
				if strings.Contains(err.Error(), "Error getting") {
					return vm, "notfound", nil
				}
				return vm, "", err
			}
			log.Printf("VM %v is currently in state %v and in LCM state %v", vm.ID, vmState, vmLcmState)
			if vmState == 3 && vmLcmState == 3 {
				return vm, "running", nil
			} else if vmState == 6 {
				return vm, "done", nil
			} else if vmState == 2 && vmLcmState == 0 {
				return vm, "hold", nil
			} else if vmState == 3 && vmLcmState == 36 {
				vmerr, _ := vm.UserTemplate.Get(vmk.Error)
				return vm, "boot_failure", fmt.Errorf("VM ID %s entered fail state, error message: %s", d.Id(), vmerr)
			} else if vmState == 3 && vmLcmState == 39 {
				vmerr, _ := vm.UserTemplate.Get(vmk.Error)
				return vm, "prolog_failure", fmt.Errorf("VM ID %s entered fail state, error message: %s", d.Id(), vmerr)
			} else if vmState == 3 && vmLcmState == 40 {
				vmerr, _ := vm.UserTemplate.Get(vmk.Error)
				return vm, "epilog_failure", fmt.Errorf("VM ID %s entered fail state, error message: %s", d.Id(), vmerr)
			} else {
				return vm, "anythingelse", nil
			}
		},
		Timeout:    3 * time.Minute,
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
	graphics := d.Get("graphics").(*schema.Set).List()
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
	os := d.Get("os").(*schema.Set).List()
	//vmos := make([]vmOs, len(os))
	for i := 0; i < len(os); i++ {
		osconfig := os[i].(map[string]interface{})
		tpl.AddOS(vmk.Arch, osconfig["arch"].(string))
		tpl.AddOS(vmk.Boot, osconfig["boot"].(string))
	}

	//Generate VM Group definition
	vmgroup := d.Get("vmgroup").(*schema.Set).List()
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
