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
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

var vmDiskUpdateReadyStates = []string{"RUNNING", "POWEROFF"}

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
			"cpu":      cpuSchema(),
			"vcpu":     vcpuSchema(),
			"memory":   memorySchema(),
			"context":  contextSchema(),
			"disk":     diskSchema(),
			"graphics": graphicsSchema(),
			"nic":      nicSchema(),
			"os":       osSchema(),
			"vmgroup":  vmGroupSchema(),
			"tags":     tagsSchema(),
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
	if v := d.Get("template_id").(int); v != -1 {
		// if template id is set, instantiate a VM from this template
		tc := controller.Template(v)

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

	return resourceOpennebulaVirtualMachineRead(d, meta)
}

func resourceOpennebulaVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
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

		// wait for the VM to be ready for attach operations
		_, err = waitForVMState(vmc, timeout, vmDiskUpdateReadyStates...)
		if err != nil {
			return fmt.Errorf(
				"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(vmDiskUpdateReadyStates, " "), err)
		}

		// get the list of disks ID to detach
		toDetach := disksConfigDiff(attachedDisksCfg, newDisksCfg)

		// Detach the disks
		for _, diskConfig := range toDetach {

			diskID := diskConfig["disk_id"].(int)

			err := vmDiskDetach(vmc, timeout, diskID)
			if err != nil {
				return fmt.Errorf("vm disk detach: %s", err)

			}
		}

		// get the list of disks to attach
		toAttach := disksConfigDiff(newDisksCfg, attachedDisksCfg)

		// Attach the disks
		for _, diskConfig := range toAttach {

			imageID := diskConfig["image_id"].(int)
			diskTpl := makeDiskVector(map[string]interface{}{
				"image_id": imageID,
				"target":   diskConfig["target"],
			})

			err := vmDiskAttach(vmc, timeout, diskTpl)
			if err != nil {
				return fmt.Errorf("vm disk attach: %s", err)
			}
		}
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
