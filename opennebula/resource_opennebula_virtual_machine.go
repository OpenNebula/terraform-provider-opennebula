package opennebula

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	errs "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
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
