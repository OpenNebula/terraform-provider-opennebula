package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

var vrInstancePairingKey = "TMP_TF_RESOURCE_ID"

func resourceOpennebulaVirtualRouterInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualRouterInstanceCreate,
		ReadContext:   resourceOpennebulaVirtualRouterInstanceRead,
		Exists:        resourceOpennebulaVirtualRouterInstanceExists,
		UpdateContext: resourceOpennebulaVirtualRouterInstanceUpdate,
		DeleteContext: resourceOpennebulaVirtualRouterInstanceDelete,
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
				"virtual_router_id": {
					Type:        schema.TypeInt,
					Required:    true,
					Description: "Identifier of the parent virtual router ressource",
				},
			},
		),
	}
}

func resourceOpennebulaVirtualRouterInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller

	vRouterID := d.Get("virtual_router_id").(int)

	// avoid creation of multiple NICs and instances at the same time
	nicKey := &SubResourceKey{
		Type:    "virtual_router",
		ID:      vRouterID,
		SubType: "nic",
	}
	config.mutex.RLock(nicKey)
	defer config.mutex.RUnlock(nicKey)

	//Call one.template.instantiate only if template_id is defined
	//otherwise use one.vm.allocate
	var err error
	var vmID int

	// retrieve the template ID from the virtual router resource
	vrc := controller.VirtualRouter(vRouterID)
	vrInfos, err := vrc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve virtual router informations",
			Detail:   err.Error(),
		})
		return diags
	}

	templateID, err := vrInfos.Template.GetInt("TEMPLATE_ID")
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve the template ID",
			Detail:   fmt.Sprintf("can't retrieve TEMPLATE_ID tag from virtual router (ID:%d)", vRouterID),
		})
		return diags
	}

	// check the template: it should exists and should be a virtual router instance template
	tpl, err := controller.Template(templateID).Info(true, false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve virtual router instance template informations",
			Detail:   err.Error(),
		})
		return diags
	}

	vrouter, _ := tpl.Template.Get("VROUTER")
	if vrouter != "YES" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Misconfigured template for the virtual router instance",
			Detail:   fmt.Sprintf("the template (ID:%d) doesn't contains the VROUTER=YES tag", templateID),
		})
		return diags
	}

	tplContext, _ := tpl.Template.GetVector(vmk.ContextVec)
	vmTpl, err := generateVm(d, tplContext)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate description",
			Detail:   err.Error(),
		})
		return diags
	}

	// The method instantiate for the virtual router doesn't returns the ID of the created VM,
	// we need to retrieve the VM ID ourselves

	// retrieve all the VM IDs associated to the virtual router
	vmsIDsSet := schema.NewSet(schema.HashInt, []interface{}{})
	for _, id := range vrInfos.VMs.ID {
		vmsIDsSet.Add(id)
	}

	// Instantiate a single virtual router instance template
	tmpProviderID, err := uuid.GenerateUUID()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate an temporary ID for the virtual router instance",
			Detail:   err.Error(),
		})
		return diags
	}

	// add instance pairing key
	vmTpl.AddPair(vrInstancePairingKey, tmpProviderID)

	vmDef := vmTpl.String()
	log.Printf("[INFO] VM definition: %s", vmDef)
	_, err = vrc.Instantiate(1, templateID, d.Get("name").(string), d.Get("pending").(bool), vmDef)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create",
			Detail:   err.Error(),
		})
		return diags
	}

	vrInfos, err = vrc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   err.Error(),
		})
		return diags
	}

	// retrieve all the VM IDs associated to the virtual router
	newVMIDs := make([]string, 0, 1)
	for _, id := range vrInfos.VMs.ID {
		if vmsIDsSet.Contains(id) {
			continue
		}
		newVMIDs = append(newVMIDs, fmt.Sprint(id))
	}

	// Retrieve light virtual router instance datas
	vmSet, err := controller.VMs().InfoSet(strings.Join(newVMIDs, ","), true)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve the virtual router instances informations",
			Detail:   err.Error(),
		})
		return diags
	}

	found := false
	for _, vmInfo := range vmSet.VMs {

		tmpID, err := vmInfo.UserTemplate.GetStr(vrInstancePairingKey)
		if err != nil {
			continue
		}

		if tmpID == tmpProviderID {
			vmID = vmInfo.ID
			found = true
			break
		}
	}
	if !found {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve the created instance",
			Detail:   fmt.Sprintf("VM with template tag %s=%s not found", vrInstancePairingKey, tmpProviderID),
		})
		return diags
	}

	// retrieve user template
	vmInfos, err := controller.VM(vmID).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   err.Error(),
		})
		return diags
	}

	// remove temporary key from template
	tmpTemplate := vmInfos.UserTemplate
	tmpTemplate.Del(vrInstancePairingKey)

	err = controller.VM(vmID).Update(tmpTemplate.String(), parameters.Replace)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to update content",
			Detail:   fmt.Sprintf("Unable to remove temporary tag %s from VM (ID:%d)", vrInstancePairingKey, vmID),
		})
		return diags
	}

	log.Printf("[DEBUG] Virtual router instance instantiated")

	d.SetId(fmt.Sprintf("%v", vmID))
	vmc := controller.VM(vmID)

	final := NewVMLCMState(vm.Running)
	if d.Get("pending").(bool) {
		final = NewVMState(vm.State(vm.Running))
	}

	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute

	finalStrs := final.ToStrings()

	_, err = waitForVMStates(ctx, vmc, resource.StateChangeConf{
		Pending:    vmCreateTransientStates.ToStrings(),
		Target:     finalStrs,
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 2 * time.Second,
	})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to wait instance to be in %s state", finalStrs),
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
				Detail:   err.Error(),
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
				Detail:   err.Error(),
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

	// Customize read step to process disk from template in a different way.
	// The goal is to avoid diffs that would trigger unwanted disk update.
	if templateID != -1 {

		flattenDiskFunc := flattenVMDisk

		if len(d.Get("disk").([]interface{})) == 0 {
			// if no disks overrides those from templates
			flattenDiskFunc = flattenVMTemplateDisk
		} else {
			d.Set("template_disk", []interface{}{})
		}

		diags = resourceOpennebulaVirtualMachineReadCustom(ctx, d, meta, func(ctx context.Context, d *schema.ResourceData, vmInfos *vm.VM) diag.Diagnostics {

			err := flattenDiskFunc(d, &vmInfos.Template)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to flatten disks",
					Detail:   err.Error(),
				})
				return diags
			}

			return nil
		})

		if len(diags) > 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to read virtual router instance informations",
			})
		}

		return diags
	}

	d.Set("template_disk", []interface{}{})

	return resourceOpennebulaVirtualRouterInstanceRead(ctx, d, meta)
}

func resourceOpennebulaVirtualRouterInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	diags := resourceOpennebulaVirtualMachineReadCustom(ctx, d, meta, func(ctx context.Context, d *schema.ResourceData, vmInfos *vm.VM) diag.Diagnostics {

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

		return nil
	})

	if len(diags) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to read",
		})
	}

	return diags
}

func resourceOpennebulaVirtualRouterInstanceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	vmID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.VM(int(vmID)).Info(false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaVirtualRouterInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	diags := resourceOpennebulaVirtualMachineUpdateCustom(ctx, d, meta, nil)
	if len(diags) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to update",
		})
	}

	return resourceOpennebulaVirtualRouterInstanceRead(ctx, d, meta)
}

func resourceOpennebulaVirtualRouterInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	diags := resourceOpennebulaVirtualMachineDelete(ctx, d, meta)
	if len(diags) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
		})
	}

	return diags

}
