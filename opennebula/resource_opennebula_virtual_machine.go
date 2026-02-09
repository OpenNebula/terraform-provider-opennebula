package opennebula

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

type NICUpdates struct {
	newNics                 []any //new nics to attach
	deletedNics             []any //nics to delete permanently
	nicsReattachedUpdated   []any //nics to reattach because they were updated
	nicsReattachedReordered []any //nics to reattach for keeping nic order
	allNicsToAttach         []any //new + reattached nics
	allNicsToDetach         []any //deleted + reattached nics
}

func (nu NICUpdates) getNICsIDs(nics []any) ([]int, error) {
	ids := make([]int, 0, len(nics))
	for _, nic := range nics {
		nicMap, ok := nic.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid nic format: %v", nic)
		}
		nicID, ok := nicMap["nic_id"].(int)
		if !ok {
			//All NICs to attach should already have a nic_id set
			return nil, fmt.Errorf("nic_id field not found in nic: %v", nic)
		}
		ids = append(ids, nicID)
	}
	return ids, nil
}

// return a map with NIC alias IDs from the nics to be recreated
func (nu NICUpdates) getRecreatedNICDependantAliasIDsMaps() (map[int][]int, error) {
	recreatedNics := append(nu.nicsReattachedReordered, nu.nicsReattachedUpdated...)
	return getDependantNICAliasesIDsMap(recreatedNics)
}

func (nu NICUpdates) getNICsToDetachIds() ([]int, error) {
	return nu.getNICsIDs(nu.deletedNics)
}

var (
	vmDiskOnChangeValues = []string{"RECREATE", "SWAP"}

	defaultVMTimeoutMin = 20
	defaultVMTimeout    = time.Duration(defaultVMTimeoutMin) * time.Minute
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
				"ip": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Primary IPv4 address assigned by OpenNebula",
				},
				"ip6": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Primary IPv6 address assigned by OpenNebula",
				},
				"ip6_ula": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Unique local IPv6 address assigned by OpenNebula",
				},
				"ip6_global": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Global IPv6 address assigned by OpenNebula",
				},
				"ip6_link": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Link-local IPv6 address assigned by OpenNebula.",
				},
				"nic":       nicVMSchema(),
				"nic_alias": nicAliasVMSchema(),
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
				"template_nic":       templateNICVMSchema(),
				"template_nic_alias": templateNICAliasVMSchema(),
			},
		),
	}
}

func nicAndAliasCommonComputedFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"nic_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"computed_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_ip": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_ip6": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_ip6_ula": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_ip6_global": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_ip6_link": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_mac": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_method": {
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
		"computed_gateway": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_dns": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func nicComputedVMFields() map[string]*schema.Schema {
	return mergeSchemas(
		nicAndAliasCommonComputedFields(),
		map[string]*schema.Schema{
			"computed_alias_ids": {
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
			"computed_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
		})
}

func nicAliasComputedVMFields() map[string]*schema.Schema {
	return mergeSchemas(
		nicAndAliasCommonComputedFields(),
		map[string]*schema.Schema{
			"computed_alias_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"computed_parent_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"computed_parent": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"computed_network_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"computed_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
		})
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

func templateNICAliasVMSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Network adapter(s) assigned to the Virtual Machine via a template",
		Elem: &schema.Resource{
			Schema: mergeSchemas(nicAliasComputedVMFields(), map[string]*schema.Schema{
				"network": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"network_id": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				"parent": {
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

func nicAliasVMFields() map[string]*schema.Schema {
	return mergeSchemas(nicAliasFields(), nicAliasComputedVMFields())
}

func nicAliasVMSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Definition of network adapter(s) assigned to the Virtual Machine",
		Elem: &schema.Resource{
			Schema: nicAliasVMFields(),
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
		"computed_dev_prefix": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_cache": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_discard": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"computed_io": {
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

func getVirtualMachineController(d *schema.ResourceData, meta interface{}) (*goca.VMController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	vmID, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.VM(int(vmID)), nil
}

func changeVmGroup(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller
	var gid int

	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	group := d.Get("group").(string)
	gid, err = controller.Groups().ByName(group)
	if err != nil {
		return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
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
				Detail:   fmt.Sprintf("VM Template (ID: %d): %s", templateID, err),
			})
			return diags
		}

		// customize template except for memory and cpu.
		vmTpl, err := generateVm(d, config, &tpl.Template)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate description",
				Detail:   err.Error(),
			})
			return diags
		}

		addNICs(d, vmTpl)
		addNICAliases(d, vmTpl)

		log.Printf("[DEBUG] VM template: %s", vmTpl.String())

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

		// save inherited template content
		inheritedTags := make(map[string]interface{})
		for _, e := range tpl.Template.Elements {
			if pair, ok := e.(*dyn.Pair); ok {
				inheritedTags[pair.Key()] = pair.Value
			}
		}

		d.Set("template_tags", inheritedTags)

		// save inherited template sections names
		inheritedSectionsNames := make(map[string]interface{})
		for _, e := range tpl.Template.Elements {
			if vec, ok := e.(*dyn.Vector); ok {
				inheritedSectionsNames[vec.Key()] = ""
			}
		}

		d.Set("template_section_names", inheritedSectionsNames)

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

		vmTpl, err := generateVm(d, config, nil)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate description",
				Detail:   err.Error(),
			})
			return diags
		}

		addNICs(d, vmTpl)
		addNICAliases(d, vmTpl)

		log.Printf("[DEBUG] VM template: %s", vmTpl.String())

		// Create VM not in pending state
		vmID, err = controller.VMs().Create(vmTpl.String(), d.Get("pending").(bool))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create the VM",
				Detail:   err.Error(),
			})
			return diags
		}

		d.Set("template_tags", map[string]interface{}{})
		d.Set("template_section_names", map[string]interface{}{})
	}

	d.SetId(fmt.Sprintf("%v", vmID))
	vmc := controller.VM(vmID)

	final := NewVMLCMState(vm.Running)
	if d.Get("pending").(bool) {
		final = NewVMState(vm.Hold)
	}

	// wait for the VM to be started and ready
	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultVMTimeout {
		timeout = d.Timeout(schema.TimeoutCreate)
	}

	stateConf := NewVMStateConf(timeout,
		vmCreateTransientStates.ToStrings(),
		final.ToStrings(),
	)
	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to wait virtual machine to be in %s state", strings.Join(final.ToStrings(), ",")),
			Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// finalize the VM configuration
	//Set the permissions on the VM if it was defined, otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = vmc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",

				Detail: fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("group") != "" {
		err = changeVmGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",

				Detail: fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = vmc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	// Customize read step to process disk and NIC from template in a different way.
	// The goal is to avoid diffs that would trigger unwanted disk/NIC update.
	if templateID != -1 {

		flattenDiskFunc := flattenVMDisk
		flattenNICFunc := flattenVMNIC
		flattenNICAliasFunc := flattenVMNICAliases

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

		if len(d.Get("nic_alias").([]interface{})) == 0 {
			// if no nics overrides those from templates
			flattenNICAliasFunc = flattenVMTemplateNICAliases
		} else {
			d.Set("template_nic_alias", []interface{}{})
		}

		return resourceOpennebulaVirtualMachineReadCustom(ctx, d, meta, func(ctx context.Context, d *schema.ResourceData, vmInfos *vm.VM) diag.Diagnostics {

			err := flattenDiskFunc(d, &vmInfos.Template)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to flatten disks",
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

			err = flattenNICFunc(d, &vmInfos.Template)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to flatten NICs",
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

			err = flattenNICAliasFunc(d, &vmInfos.Template)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to flatten NIC Aliases",
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
			return nil
		})

	}

	d.Set("template_nic", []interface{}{})
	d.Set("template_disk", []interface{}{})
	d.Set("template_nic_alias", []interface{}{})

	contextWo := getContextWo(d)
	if err := d.Set("context_wo_hash", computeContextWoHash(contextWo)); err != nil {
		log.Printf("[WARN] Failed to set context_wo_hash: %s", err)
	}

	return resourceOpennebulaVirtualMachineRead(ctx, d, meta)
}

func resourceOpennebulaVirtualMachineReadCustom(ctx context.Context, d *schema.ResourceData, meta interface{}, customVM customVMFunc) diag.Diagnostics {

	var diags diag.Diagnostics

	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the virtual machine controller",
			Detail:   err.Error(),
		})
		return diags

	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vmInfo, err := vmc.Info(false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual machine %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%v", vmInfo.ID))
	d.Set("name", vmInfo.Name)
	d.Set("uid", vmInfo.UID)
	d.Set("gid", vmInfo.GID)
	d.Set("uname", vmInfo.UName)
	d.Set("gname", vmInfo.GName)
	d.Set("state", vmInfo.StateRaw)
	d.Set("lcmstate", vmInfo.LCMStateRaw)
	if vm.State(vmInfo.StateRaw) == vm.Done {
		log.Printf("[WARN] Replacing virtual machine %s (id: %s) because VM is 'Done'; ", d.Get("name"), d.Id())
		d.SetId("")
		return nil
	}
	//TODO fix this:
	err = d.Set("permissions", permissionsUnixString(*vmInfo.Permissions))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set attribute",
			Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if customVM != nil {
		customDiags := customVM(ctx, d, vmInfo)
		if len(customDiags) > 0 {
			return customDiags
		}
	}

	var inheritedVectors map[string]interface{}
	inheritedVectorsIf := d.Get("template_section_names")
	if inheritedVectorsIf != nil {
		inheritedVectors = inheritedVectorsIf.(map[string]interface{})
	}
	err = flattenTemplate(d, inheritedVectors, &vmInfo.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten",
			Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	var inheritedTags map[string]interface{}
	inheritedTagsIf := d.Get("template_tags")
	if inheritedTagsIf != nil {
		inheritedTags = inheritedTagsIf.(map[string]interface{})
	}

	flattenDiags := flattenVMUserTemplate(d, meta, inheritedTags, &vmInfo.UserTemplate.Template)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
	}

	if vmInfo.LockInfos != nil {
		d.Set("lock", LockLevelToString(vmInfo.LockInfos.Locked))
	}

	return diags
}

func resourceOpennebulaVirtualMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceOpennebulaVirtualMachineReadCustom(ctx, d, meta, func(ctx context.Context, d *schema.ResourceData, vmInfos *vm.VM) diag.Diagnostics {

		var diags diag.Diagnostics

		// NOTE: The template_id attribute is not defined for Virtual Router instances (VMs).
		if _, ok := d.GetOk("template_id"); ok {
			// read template ID from which the VM was created
			templateID, _ := vmInfos.Template.GetInt("TEMPLATE_ID")
			d.Set("template_id", templateID)

			if _, ok := d.GetOk("template_nic"); !ok {
				d.Set("template_nic", []interface{}{})
			}

			if _, ok := d.GetOk("template_nic_alias"); !ok {
				d.Set("template_nic_alias", []interface{}{})
			}
		}

		// add empty values for import
		if _, ok := d.GetOk("template_disk"); !ok {
			d.Set("template_disk", []interface{}{})
		}
		if _, ok := d.GetOk("template_tags"); !ok {
			d.Set("template_tags", map[string]interface{}{})
		}
		if _, ok := d.GetOk("template_section_names"); !ok {
			d.Set("template_section_names", map[string]interface{}{})
		}

		err := flattenVMDisk(d, &vmInfos.Template)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to flatten disks",
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		// In case of Virtual Router instances (which are just VMs) there's never anything to "flatten",
		// that's because NICs are attached with a help of dedicated resources.
		if _, ok := d.GetOk("nic"); ok {
			err = flattenVMNIC(d, &vmInfos.Template)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to flatten NICs",
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		if _, ok := d.GetOk("nic_alias"); ok {
			err = flattenVMNICAliases(d, &vmInfos.Template)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to flatten NIC aliases",
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		// Don't read context from API: it returns merged context+context_wo which
		// would leak write-only values into state and break idempotency.

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
	if len(diskConfig["dev_prefix"].(string)) > 0 {
		diskMap["dev_prefix"] = diskMap["computed_dev_prefix"]
	}
	if len(diskConfig["cache"].(string)) > 0 {
		diskMap["cache"] = diskMap["computed_cache"]
	}
	if len(diskConfig["discard"].(string)) > 0 {
		diskMap["discard"] = diskMap["computed_discard"]
	}
	if len(diskConfig["io"].(string)) > 0 {
		diskMap["io"] = diskMap["computed_io"]
	}

	return diskMap
}

func flattenDiskComputed(disk shared.Disk) map[string]interface{} {
	size, _ := disk.GetI(shared.Size)
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	volatileFormat, _ := disk.Get("FORMAT")
	dev_prefix, _ := disk.Get("DEV_PREFIX")
	cache, _ := disk.Get("CACHE")
	discard, _ := disk.Get("DISCARD")
	io, _ := disk.Get("IO")
	diskID, _ := disk.GetI(shared.DiskID)

	return map[string]interface{}{
		"disk_id":                  diskID,
		"computed_size":            size,
		"computed_target":          target,
		"computed_driver":          driver,
		"computed_volatile_format": volatileFormat,
		"computed_dev_prefix":      dev_prefix,
		"computed_cache":           cache,
		"computed_discard":         discard,
		"computed_io":              io,
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
	dev_prefix, _ := disk.Get("DEV_PREFIX")
	cache, _ := disk.Get("CACHE")
	discard, _ := disk.Get("DISCARD")
	io, _ := disk.Get("IO")
	volatileType, _ := disk.Get("TYPE")
	volatileFormat, _ := disk.Get("FORMAT")

	return emptyOrEqual(diskConfig["target"], target) &&
		emptyOrEqual(diskConfig["size"], size) &&
		emptyOrEqual(diskConfig["driver"], driver) &&
		emptyOrEqual(diskConfig["dev_prefix"], dev_prefix) &&
		emptyOrEqual(diskConfig["cache"], cache) &&
		emptyOrEqual(diskConfig["discard"], discard) &&
		emptyOrEqual(diskConfig["io"], io) &&
		emptyOrEqual(diskConfig["volatile_type"], volatileType) &&
		emptyOrEqual(diskConfig["volatile_format"], volatileFormat)
}

func matchDiskComputed(diskConfig map[string]interface{}, disk shared.Disk) bool {

	size, _ := disk.GetI(shared.Size)
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	format, _ := disk.Get("FORMAT")
	dev_prefix, _ := disk.Get("DEV_PREFIX")
	cache, _ := disk.Get("CACHE")
	discard, _ := disk.Get("DISCARD")
	io, _ := disk.Get("IO")

	return (target == diskConfig["computed_target"].(string)) &&
		(size == diskConfig["computed_size"].(int)) &&
		(driver == diskConfig["computed_driver"].(string)) &&
		(format == diskConfig["computed_volatile_format"].(string)) &&
		(dev_prefix == diskConfig["computed_dev_prefix"].(string)) &&
		(cache == diskConfig["computed_cache"].(string)) &&
		(discard == diskConfig["computed_discard"].(string)) &&
		(io == diskConfig["computed_io"].(string))
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

	err := d.Set("disk", diskList)
	if err != nil {
		return err
	}

	return nil
}

func flattenNICComputedAttributes(nic shared.NIC, ignoreSGIDs []int) map[string]interface{} {
	nicExclusiveComputedAttributes := flattenNICExclusiveComputedAttributes(nic)
	nicAndAliasCommonComputedAttributes := flattenNICAndAliasCommonComputedAttributes(nic, ignoreSGIDs)

	flattenedNICComputedAttributes := nicAndAliasCommonComputedAttributes
	for k, v := range nicExclusiveComputedAttributes {
		flattenedNICComputedAttributes[k] = v
	}
	return flattenedNICComputedAttributes
}

func flattenNICAliasComputedAttributes(nic shared.NIC, ignoreSGIDs []int) map[string]interface{} {
	nicAliasExclusiveComputedAttributes := flattenNICAliasExclusiveComputedAttributes(nic)
	nicAndAliasCommonComputedAttributes := flattenNICAndAliasCommonComputedAttributes(nic, ignoreSGIDs)

	flattenedNICAliasComputedAttributes := nicAndAliasCommonComputedAttributes
	for k, v := range nicAliasExclusiveComputedAttributes {
		flattenedNICAliasComputedAttributes[k] = v
	}
	return flattenedNICAliasComputedAttributes
}

func flattenNICExclusiveComputedAttributes(nic shared.NIC) map[string]interface{} {

	aliasIDs, _ := nic.Get(shared.AliasIDs)
	network, _ := nic.Get(shared.Network)
	physicalDevice, _ := nic.GetStr("PHYDEV")
	model, _ := nic.Get(shared.Model)
	virtioQueues, _ := nic.GetStr("VIRTIO_QUEUES")
	method, _ := nic.Get(shared.Method)

	attributeMap := map[string]interface{}{
		"computed_alias_ids":       aliasIDs,
		"network":                  network,
		"computed_physical_device": physicalDevice,
		"computed_model":           model,
		"computed_virtio_queues":   virtioQueues,
		"computed_method":          method,
	}

	return attributeMap
}

func flattenNICAliasExclusiveComputedAttributes(nic shared.NIC) map[string]interface{} {
	aliasID, _ := nic.GetI(shared.NICAliasID)
	parent, _ := nic.Get(shared.NICAliasParent)
	parentID, _ := nic.GetI(shared.NICAliasParentID)
	//network and network ID could be set or computed in NIC Alias
	networkID, _ := nic.GetI(shared.NetworkID)
	network, _ := nic.Get(shared.Network)

	attributeMap := map[string]interface{}{
		"computed_alias_id":   aliasID,
		"computed_parent":     parent,
		"computed_parent_id":  parentID,
		"computed_network_id": networkID,
		"computed_network":    network,
	}

	return attributeMap
}

func flattenNICAndAliasCommonComputedAttributes(nic shared.NIC, ignoreSGIDs []int) map[string]interface{} {
	nicID, _ := nic.ID()
	name, _ := nic.Get(shared.Name)
	sg := make([]int, 0)
	ip, _ := nic.Get(shared.IP)
	mac, _ := nic.Get(shared.MAC)
	ip6, _ := nic.Get(shared.IP6)
	ip6ULA, _ := nic.Get(shared.IP6_ULA)
	ip6Global, _ := nic.Get(shared.IP6_GLOBAL)
	ip6Link, _ := nic.Get(shared.IP6_LINK)
	gateway, _ := nic.Get(shared.Gateway)
	dns, _ := nic.Get(shared.DNS)

	securityGroupsArray, _ := nic.Get(shared.SecurityGroups)
	sgString := strings.Split(securityGroupsArray, ",")
	for _, s := range sgString {
		sgInt, _ := strconv.ParseInt(s, 10, 32)

		// OpenNebula adds default security group, we may want to avoid a diff
		ignored := false
		for _, id := range ignoreSGIDs {
			if id == int(sgInt) {
				ignored = true
			}
		}
		if ignored {
			continue
		}
		sg = append(sg, int(sgInt))
	}

	return map[string]interface{}{
		"nic_id":                   nicID,
		"computed_name":            name,
		"computed_ip":              ip,
		"computed_ip6":             ip6,
		"computed_ip6_ula":         ip6ULA,
		"computed_ip6_global":      ip6Global,
		"computed_ip6_link":        ip6Link,
		"computed_mac":             mac,
		"computed_security_groups": sg,
		"computed_gateway":         gateway,
		"computed_dns":             dns,
	}
}

// Flatten VM NIC attributes from OpenNebula API computed values
func flattenVMNICComputedAttributes(NICConfig map[string]interface{}, NIC shared.NIC) map[string]interface{} {

	NICMap := flattenNICExclusiveComputedAttributes(NIC)

	//Override the resource values with the computed ones
	if len(NICConfig["model"].(string)) > 0 {
		NICMap["model"] = NICMap["computed_model"]
	}
	if len(NICConfig["virtio_queues"].(string)) > 0 {
		NICMap["virtio_queues"] = NICMap["computed_virtio_queues"]
	}
	if len(NICConfig["physical_device"].(string)) > 0 {
		NICMap["physical_device"] = NICMap["computed_physical_device"]
	}
	if len(NICConfig["method"].(string)) > 0 {
		NICMap["method"] = NICMap["computed_method"]
	}

	networkMode, err := NIC.Get(shared.NetworkMode)
	if err == nil && networkMode == "auto" {
		NICMap["network_mode_auto"] = true
	}

	schedReqs, err := NIC.Get(shared.SchedRequirements)
	if err == nil {
		NICMap["sched_requirements"] = schedReqs
	}

	schedRank, err := NIC.Get(shared.SchedRank)
	if err == nil {
		NICMap["sched_rank"] = schedRank
	}

	//retrieve computed values for NIC and NICAlias common attributes
	attributeMap := flattenVMNICAndAliasCommonComputedAttributes(NICConfig, NIC)

	// merge values from NICMap into attributeMap
	for k, v := range NICMap {
		attributeMap[k] = v
	}

	return attributeMap
}

// Flatten VM NIC Alias attributes from OpenNebula API computed values
func flattenVMNICAliasComputedAttributes(NICConfig map[string]interface{}, NIC shared.NIC) map[string]interface{} {

	NICMap := flattenNICAliasExclusiveComputedAttributes(NIC)

	//Override the resource values with the computed ones
	if len(NICConfig["network"].(string)) > 0 {
		NICMap["network"] = NICMap["computed_network"]
	}
	if v, ok := NICConfig["network_id"].(int); ok && v > -1 {
		NICMap["network_id"] = NICMap["computed_network_id"]
	}
	if len(NICConfig["parent"].(string)) > 0 {
		NICMap["parent"] = NICMap["computed_parent"]
	}

	//retrieve computed values for NIC and NICAlias common attributes
	attributeMap := flattenVMNICAndAliasCommonComputedAttributes(NICConfig, NIC)

	// merge values from NICMap into attributeMap (already existing attributeMap values could be overridden)
	for k, v := range NICMap {
		attributeMap[k] = v
	}

	return attributeMap
}

// Flatten VM NIC and NIC Alias common attributes from OpenNebula API computed values
func flattenVMNICAndAliasCommonComputedAttributes(NICConfig map[string]interface{}, NIC shared.NIC) map[string]interface{} {
	// Flatten the NIC configuration to a map with values coming from OpenNebula API (computed values)
	NICMap := flattenNICAndAliasCommonComputedAttributes(NIC, []int{0})

	//Override the resource values with the computed ones
	if len(NICConfig["name"].(string)) > 0 {
		NICMap["name"] = NICMap["computed_name"]
	}
	if len(NICConfig["ip"].(string)) > 0 {
		NICMap["ip"] = NICMap["computed_ip"]
	}
	if len(NICConfig["ip6"].(string)) > 0 {
		NICMap["ip6"] = NICMap["computed_ip6"]
	}
	if len(NICConfig["ip6_ula"].(string)) > 0 {
		NICMap["ip6_ula"] = NICMap["computed_ip6_ula"]
	}
	if len(NICConfig["ip6_global"].(string)) > 0 {
		NICMap["ip6_global"] = NICMap["computed_ip6_global"]
	}
	if len(NICConfig["ip6_link"].(string)) > 0 {
		NICMap["ip6_link"] = NICMap["computed_ip6_link"]
	}
	if len(NICConfig["mac"].(string)) > 0 {
		NICMap["mac"] = NICMap["computed_mac"]
	}
	if len(NICConfig["security_groups"].([]interface{})) > 0 {
		NICMap["security_groups"] = NICMap["computed_security_groups"]
	}
	if len(NICConfig["gateway"].(string)) > 0 {
		NICMap["gateway"] = NICMap["computed_gateway"]
	}
	if len(NICConfig["dns"].(string)) > 0 {
		NICMap["dns"] = NICMap["computed_dns"]
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
		nicRead := flattenNICComputedAttributes(nic, nil)
		nicRead["network_id"] = networkID
		nicList = append(nicList, nicRead)

		if i == 0 {
			d.Set("ip", nicRead["computed_ip"])
			d.Set("ip6", nicRead["computed_ip6"])
			d.Set("ip6_ula", nicRead["computed_ip6_ula"])
			d.Set("ip6_global", nicRead["computed_ip6_global"])
			d.Set("ip6_link", nicRead["computed_ip6_link"])
		}
	}

	err := d.Set("template_nic", nicList)
	if err != nil {
		return err
	}

	return nil
}

// flattenVMTemplateNIC read NIC Alias that come from template when instantiating a VM
func flattenVMTemplateNICAliases(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set Nics to resource
	nicAliases := vmTemplate.GetNICAliases()
	nicAliasList := make([]interface{}, 0, len(nicAliases))

	for _, nicAlias := range nicAliases {
		networkID, _ := nicAlias.GetI(shared.NetworkID)
		network, _ := nicAlias.Get(shared.Network)
		parent, _ := nicAlias.Get(shared.NICAliasParent)

		nicAliasRead := flattenNICAliasComputedAttributes(nicAlias, nil)
		nicAliasRead["network_id"] = networkID
		nicAliasRead["network"] = network
		nicAliasRead["parent"] = parent
		nicAliasList = append(nicAliasList, nicAliasRead)
	}

	err := d.Set("template_nic_alias", nicAliasList)
	if err != nil {
		return err
	}

	return nil
}

func matchNIC(NICConfig map[string]interface{}, NIC shared.NIC) bool {

	physicalDevice, _ := NIC.GetStr("PHYDEV")
	model, _ := NIC.Get(shared.Model)
	virtioQueues, _ := NIC.GetStr("VIRTIO_QUEUES")
	schedRequirements, _ := NIC.Get(shared.SchedRequirements)
	schedRank, _ := NIC.Get(shared.SchedRank)
	networkMode, _ := NIC.Get(shared.NetworkMode)
	method, _ := NIC.Get(shared.Method)

	matchCommonFields := matchNICAndAliasCommonAttributes(NICConfig, NIC)

	return matchCommonFields &&
		emptyOrEqual(NICConfig["physical_device"], physicalDevice) &&
		emptyOrEqual(NICConfig["model"], model) &&
		emptyOrEqual(NICConfig["virtio_queues"], virtioQueues) &&
		emptyOrEqual(NICConfig["method"], method) &&
		emptyOrEqual(NICConfig["sched_requirements"], schedRequirements) &&
		emptyOrEqual(NICConfig["sched_rank"], schedRank) &&
		(NICConfig["network_mode_auto"].(bool) == false || networkMode == "auto")
}

func matchNICAlias(NICConfig map[string]interface{}, NIC shared.NIC) bool {
	parent, _ := NIC.Get(shared.NICAliasParent)
	//aliasId, _ := NIC.GetI(shared.NICAliasID)
	network, _ := NIC.Get(shared.Network)
	networkId, _ := NIC.GetI(shared.NetworkID)

	//workaround for network_id being set to -1 in case of no network set on resource
	resourceNetworkID, ok := NICConfig["network_id"].(int)
	if ok && resourceNetworkID == -1 {
		resourceNetworkID = 0
	}

	matchCommonFields := matchNICAndAliasCommonAttributes(NICConfig, NIC)

	return matchCommonFields &&
		emptyOrEqual(NICConfig["parent"], parent) &&
		emptyOrEqual(NICConfig["network"], network) &&
		emptyOrEqual(resourceNetworkID, networkId)
}

func matchNICAndAliasCommonAttributes(NICConfig map[string]interface{}, NIC shared.NIC) bool {
	name, _ := NIC.Get(shared.Name)
	ip, _ := NIC.Get(shared.IP)
	ip6, _ := NIC.Get(shared.IP6)
	ip6ULA, _ := NIC.Get(shared.IP6_ULA)
	ip6Global, _ := NIC.Get(shared.IP6_GLOBAL)
	ip6Link, _ := NIC.Get(shared.IP6_LINK)
	mac, _ := NIC.Get(shared.MAC)
	gateway, _ := NIC.Get(shared.Gateway)
	dns, _ := NIC.Get(shared.DNS)

	sgMatches := checkNICSGMatches(NICConfig, NIC)

	return sgMatches &&
		emptyOrEqual(NICConfig["name"], name) &&
		emptyOrEqual(NICConfig["ip"], ip) &&
		emptyOrEqual(NICConfig["ip6"], ip6) &&
		emptyOrEqual(NICConfig["ip6_ula"], ip6ULA) &&
		emptyOrEqual(NICConfig["ip6_global"], ip6Global) &&
		emptyOrEqual(NICConfig["ip6_link"], ip6Link) &&
		emptyOrEqual(NICConfig["mac"], mac) &&
		emptyOrEqual(NICConfig["gateway"], gateway) &&
		emptyOrEqual(NICConfig["dns"], dns)
}

func checkNICSGMatches(NICConfig map[string]interface{}, NIC shared.NIC) bool {

	if NICConfig["security_groups"] == nil || len(NICConfig["security_groups"].([]interface{})) == 0 {
		return true
	}

	securityGroupsArray, _ := NIC.Get(shared.SecurityGroups) // SGs from opennebula
	sg := strings.Split(securityGroupsArray, ",")

	sgConfig := NICConfig["security_groups"].([]interface{}) //resource SGs

	// check that sgConfig is included in sg.
	// equality is not possible since OpenNebula adds the default security group 0
	for i := 0; i < len(sgConfig); i++ {
		match := false

		for j := 0; j < len(sg); j++ {

			sgInt, err := strconv.ParseInt(sg[j], 10, 0)
			if err != nil {
				return false
			}

			if int(sgInt) != sgConfig[i].(int) {
				continue
			}
			match = true
			break
		}
		if !match {
			return false
		}
	}
	return true
}

func matchNICAndAliasCommonComputedAttributes(NICConfig map[string]interface{}, NIC shared.NIC) bool {
	name, _ := NIC.Get(shared.Name)
	ip, _ := NIC.Get(shared.IP)
	ip6, _ := NIC.Get(shared.IP6)
	ip6ULA, _ := NIC.Get(shared.IP6_ULA)
	ip6Global, _ := NIC.Get(shared.IP6_GLOBAL)
	ip6Link, _ := NIC.Get(shared.IP6_LINK)
	mac, _ := NIC.Get(shared.MAC)
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

	gateway, _ := NIC.Get(shared.Gateway)
	dns, _ := NIC.Get(shared.DNS)

	return name == NICConfig["computed_name"] &&
		ip == NICConfig["computed_ip"].(string) &&
		ip6 == NICConfig["computed_ip6"].(string) &&
		ip6ULA == NICConfig["computed_ip6_ula"].(string) &&
		ip6Global == NICConfig["computed_ip6_global"].(string) &&
		ip6Link == NICConfig["computed_ip6_link"].(string) &&
		mac == NICConfig["computed_mac"].(string) &&
		gateway == NICConfig["computed_gateway"].(string) &&
		dns == NICConfig["computed_dns"].(string)
}

func matchNICComputed(NICConfig map[string]interface{}, NIC shared.NIC) bool {
	physicalDevice, _ := NIC.GetStr("PHYDEV")
	model, _ := NIC.Get(shared.Model)
	virtioQueues, _ := NIC.GetStr("VIRTIO_QUEUES")
	method, _ := NIC.Get(shared.Method)

	matchCommon := matchNICAndAliasCommonComputedAttributes(NICConfig, NIC)

	return matchCommon &&
		physicalDevice == NICConfig["computed_physical_device"].(string) &&
		model == NICConfig["computed_model"].(string) &&
		virtioQueues == NICConfig["computed_virtio_queues"].(string) &&
		method == NICConfig["computed_method"].(string)
}

func matchNICAliasComputed(NICConfig map[string]interface{}, NIC shared.NIC) bool {
	parent, _ := NIC.Get(shared.NICAliasParent)
	parentId, _ := NIC.GetI(shared.NICAliasParentID)

	matchCommon := matchNICAndAliasCommonComputedAttributes(NICConfig, NIC)

	return matchCommon &&
		parent == NICConfig["computed_parent"].(string) &&
		parentId == NICConfig["computed_parent_id"].(int)
}

// flattenVMNIC is similar to flattenNIC but deal with computed_* attributes
// this is a temporary solution until we can use nested attributes marked computed and optional
func flattenVMNIC(d *schema.ResourceData, vmTemplate *vm.Template) error {

	// Set Nics to resource
	nics := vmTemplate.GetNICs()                //nics read from opennebula
	nicsConfigs := d.Get("nic").([]interface{}) //nics from resource

	nicList := make([]interface{}, 0, len(nics))

	for i, nic := range nics {

		// exclude NIC listed in template_nic
		if isNicInTemplate(d, nic) {
			continue
		}

		matchingNicConfig, match, err := findMatchingNICConfig(nicsConfigs, nic, matchNIC)
		if err != nil {
			return fmt.Errorf("could not find matching NIC config: %v", err)
		}

		if !match {
			ID, _ := nic.ID()
			log.Printf("[WARN] Configuration for NIC ID: %d not found.", ID)
			continue
		}

		nicMap := flattenVMNICComputedAttributes(matchingNicConfig, nic)

		networkIDCfg := matchingNicConfig["network_id"].(int)
		if networkIDCfg == -1 {
			nicMap["network_id"] = -1
		} else {
			networkID, _ := nic.GetI(shared.NetworkID)
			nicMap["network_id"] = networkID
		}

		nicList = append(nicList, nicMap)

		// Set the first NIC's IPs to VM resource
		if i == 0 {
			d.Set("ip", nicMap["computed_ip"])
			d.Set("ip6", nicMap["computed_ip6"])
			d.Set("ip6_ula", nicMap["computed_ip6_ula"])
			d.Set("ip6_global", nicMap["computed_ip6_global"])
			d.Set("ip6_link", nicMap["computed_ip6_link"])
		}
	}

	err := d.Set("nic", nicList)
	if err != nil {
		return err
	}

	return nil
}

func isNicInTemplate(d *schema.ResourceData, nic shared.NIC) bool {
	tplNICConfigs := d.Get("template_nic").([]interface{})
	for _, tplNICConfigIf := range tplNICConfigs {
		tplNICConfig := tplNICConfigIf.(map[string]interface{})
		if matchNICComputed(tplNICConfig, nic) {
			return true
		}
	}
	return false
}

func flattenVMNICAliases(d *schema.ResourceData, vmTemplate *vm.Template) error {

	remoteNicAliases := vmTemplate.GetNICAliases()       //nic_alias from opennebula
	resourceConfig := d.Get("nic_alias").([]interface{}) //nics from resource

	nicAliasList := make([]interface{}, 0, len(remoteNicAliases))
	for _, remoteNicAlias := range remoteNicAliases {

		// exclude NIC Alias listed in template_nic_alias
		if isNicAliasInTemplate(d, remoteNicAlias) {
			continue
		}

		matchingNicAliasConfig, match, err := findMatchingNICConfig(resourceConfig, remoteNicAlias, matchNICAlias)
		if err != nil {
			return fmt.Errorf("could not find matching NIC Alias config: %v", err)
		}

		if !match {
			nicAliasName, _ := remoteNicAlias.Get(shared.Name)
			log.Printf("[WARN] Configuration for NIC Alias: %s not found.", nicAliasName)
			continue
		}

		nicAliasMap := flattenVMNICAliasComputedAttributes(matchingNicAliasConfig, remoteNicAlias)

		//This workaround is needed because SDK2 does not work well with nested
		// attributes marked computed and optional, it sets always a default value
		// of 0, so we override the defaulte value to -1 and set it to -1
		// in the state file for not having problems with the diff.
		networkIDCfg := matchingNicAliasConfig["network_id"].(int)
		if networkIDCfg == -1 {
			nicAliasMap["network_id"] = -1
		} else {
			networkID, _ := remoteNicAlias.GetI(shared.NetworkID)
			nicAliasMap["network_id"] = networkID
		}

		networkCfg := matchingNicAliasConfig["network"].(string)
		if len(networkCfg) == 0 {
			nicAliasMap["network"] = networkCfg
		} else {
			network, _ := remoteNicAlias.Get(shared.Network)
			nicAliasMap["network"] = network
		}

		nicAliasList = append(nicAliasList, nicAliasMap)
	}

	err := d.Set("nic_alias", nicAliasList)
	if err != nil {
		return err
	}
	return nil
}

func isNicAliasInTemplate(d *schema.ResourceData, nicAlias shared.NIC) bool {
	tplNICAliasConfigs := d.Get("template_nic_alias").([]interface{})
	for _, tplNICAliasConfigIf := range tplNICAliasConfigs {
		tplNICAliasConfig := tplNICAliasConfigIf.(map[string]interface{})
		if matchNICAliasComputed(tplNICAliasConfig, nicAlias) {
			return true
		}
	}
	return false
}

func findMatchingNICConfig(nicConfigList []interface{}, nic shared.NIC,
	matchFunction func(NICConfig map[string]interface{}, NIC shared.NIC) bool) (map[string]interface{}, bool, error) {

	for _, nicConfigElement := range nicConfigList {
		nicConfig, ok := nicConfigElement.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("invalid NIC configuration element: %v", nicConfigElement)
		}
		if matchFunction(nicConfig, nic) {
			return nicConfig, true, nil
		}
	}

	return nil, false, nil
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

	if d.HasChange("nic") || d.HasChange("nic_alias") {
		err := updateNICAndAliases(ctx, d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update NIC and NIC Aliases",
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return nil
}

func updateNICAndAliases(ctx context.Context, d *schema.ResourceData, meta any) error {
	log.Printf("[DEBUG] Updating NIC and NIC Aliases for VM ID: %s", d.Id())

	var err error
	nicUpdates := NICUpdates{}
	nicAliasUpdates := NICUpdates{}

	if d.HasChange("nic") {
		nicUpdates, err = getNICUpdates(d)
		if err != nil {
			return fmt.Errorf("failed to retrieve NIC changes: %s", err)
		}
	}

	// Retrieve dependant NIC Aliases that need to be recreated
	dependantNicAliasMap, err := nicUpdates.getRecreatedNICDependantAliasIDsMaps()
	if err != nil {
		return fmt.Errorf("failed to retrieve dependant NIC Aliases: %s", err)
	}

	if d.HasChange("nic_alias") || len(dependantNicAliasMap) > 0 {
		nicAliasUpdates, err = getNICAliasUpdates(d, nicUpdates)
		if err != nil {
			return fmt.Errorf("failed to retrieve NIC Alias changes: %s", err)
		}
	}

	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultVMTimeout {
		timeout = d.Timeout(schema.TimeoutUpdate)
	}

	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return fmt.Errorf("failed to retrieve Virtual Machine Controller: %s", err)
	}

	err = detachNICAliases(ctx, vmc, timeout, nicAliasUpdates)
	if err != nil {
		return fmt.Errorf("failed to detach NIC Aliases: %s", err)
	}

	err = detachNICs(ctx, vmc, timeout, nicUpdates, nicAliasUpdates)
	if err != nil {
		return fmt.Errorf("failed to detach NICs: %s", err)
	}

	err = attachNICs(ctx, vmc, timeout, nicUpdates)
	if err != nil {
		return fmt.Errorf("failed to attach NIC Aliases: %s", err)
	}

	err = attachNICAliases(ctx, vmc, timeout, nicAliasUpdates)
	if err != nil {
		return fmt.Errorf("failed to attach NICs: %s", err)
	}

	return nil
}

func retrieveDependantNICAliasesToRecreate(nicAliasList []any, nicUpdates NICUpdates) ([]any, error) {
	// Retrieve dependant NIC Aliases that need to be recreated
	dependantNicAliasMap, err := nicUpdates.getRecreatedNICDependantAliasIDsMaps()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve dependant NIC Aliases: %s", err)
	}

	dependantNicAliasToRecreate := []any{}
	for _, nicAliasIDs := range dependantNicAliasMap {
		if len(nicAliasIDs) > 0 {
			for _, nicAliasID := range nicAliasIDs {
				nicAlias, err := findNICAliasByID(nicAliasList, nicAliasID)
				if err != nil {
					return nil, fmt.Errorf("failed to find NICAlias by ID %d: %s", nicAliasID, err)
				}
				dependantNicAliasToRecreate = append(dependantNicAliasToRecreate, nicAlias)
			}
		}
	}

	return dependantNicAliasToRecreate, nil
}

func findNICAliasByID(nicAliasList []any, nicAliasID int) (map[string]any, error) {
	for _, nicAlias := range nicAliasList {
		nicAliasMap, ok := nicAlias.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid NIC Alias format in resource data")
		}
		if nicAliasMap["nic_id"] == nicAliasID {
			return nicAliasMap, nil
		}
	}
	return nil, fmt.Errorf("NIC Alias with ID %d not found", nicAliasID)
}

func resourceOpennebulaVirtualMachineUpdateCustom(ctx context.Context, d *schema.ResourceData, meta interface{}, customFunc customFunc) diag.Diagnostics {

	var diags diag.Diagnostics

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the virtual machine controller",
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
			Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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

	if updateTemplate(d, tpl) {
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
			key := strings.ToUpper(k)
			tpl.Del(key)
			tpl.AddPair(key, v)
		}

		update = true
	}

	if d.HasChange("tags_all") {
		oldTagsAllIf, newTagsAllIf := d.GetChange("tags_all")
		oldTagsAll := oldTagsAllIf.(map[string]interface{})
		newTagsAll := newTagsAllIf.(map[string]interface{})

		tags := d.Get("tags").(map[string]interface{})

		// delete tags
		for k, _ := range oldTagsAll {
			_, ok := newTagsAll[k]
			if ok {
				continue
			}
			tpl.Del(strings.ToUpper(k))
		}

		// reapply all default tags that were neither applied nor overriden via tags section
		for k, v := range newTagsAll {
			_, ok := tags[k]
			if ok {
				continue
			}

			key := strings.ToUpper(k)
			tpl.Del(key)
			tpl.AddPair(key, v)
		}

		update = true
	}

	if update {
		err = vmc.Update(tpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
		if timeout == defaultVMTimeout {
			timeout = d.Timeout(schema.TimeoutCreate)
		}

		finalStrs := NewVMLCMState(vm.Running).ToStrings()
		stateConf := NewVMUpdateStateConf(timeout,
			[]string{},
			finalStrs,
		)
		_, err = waitForVMStates(ctx, vmc, stateConf)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to wait virtual machine to be in %s state", strings.Join(finalStrs, ",")),
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
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
	for _, name := range []string{"OS", "FEATURES", "INPUT", "GRAPHICS", "RAW", "CONTEXT", "CPU_MODEL"} {
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
						Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
						Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
		}
	}

	// context_wo changes are tracked via context_wo_hash (set by CustomizeDiff).
	contextChanged := d.HasChange("context")
	contextWoChanged := d.HasChange("context_wo_hash")

	if contextChanged || contextWoChanged {

		updateConf = true

		context := d.Get("context").(map[string]interface{})
		contextWo := getContextWo(d)

		log.Printf("[DEBUG] Update context (contextChanged=%v, contextWoChanged=%v, context=%d vars, context_wo=%d vars)",
			contextChanged, contextWoChanged, len(context), len(contextWo))

		old, new := d.GetChange("context")
		appliedContext := old.(map[string]interface{})
		newContext := new.(map[string]interface{})

		if len(newContext) == 0 && len(contextWo) == 0 {
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

				// Add write-only elements (always applied, no prior state)
				for key, value := range contextWo {
					keyUp := strings.ToUpper(key)
					contextVec.AddPair(keyUp, value)
				}

			} else {
				// Merge both context maps for update
				mergedApplied := make(map[string]interface{})
				mergedNew := make(map[string]interface{})

				for k, v := range appliedContext {
					mergedApplied[strings.ToUpper(k)] = v
				}

				for k, v := range newContext {
					mergedNew[strings.ToUpper(k)] = v
				}
				for k, v := range contextWo {
					mergedNew[strings.ToUpper(k)] = v
				}

				updateVMTemplateVec(tpl, "CONTEXT", mergedApplied, mergedNew)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to update CONTEXT vector",
						Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
		}

		// context_wo_hash is already updated by CustomizeDiff via SetNew
	}

	if d.HasChange("raw") {
		updateRaw(d, &tpl.Template)
		updateConf = true
	}

	if d.HasChange("cpumodel") {
		tpl.Del("CPU_MODEL")
		cpumodel := d.Get("cpumodel").([]interface{})

		for i := 0; i < len(cpumodel); i++ {
			cpumodelConfig := cpumodel[i].(map[string]interface{})
			tpl.CPUModel(cpumodelConfig["model"].(string))
		}

		updateConf = true
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
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

			// wait for the VM to be powered off
			// RUNNING state is added to transient one in case of slow cloud
			transient := vmPowerOffTransientStates
			transient.LCMs = append(transient.LCMs, vm.Running)
			stateConf := NewVMUpdateStateConf(timeout,
				transient.ToStrings(),
				NewVMState(vm.Poweroff).ToStrings(),
			)

			_, err = waitForVMStates(ctx, vmc, stateConf)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to wait virtual machine to be in POWEROFF state",
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		// wait for the VM to be back in POWEROFF state
		transientStrs := NewVMLCMState(vm.HotplugResize).ToStrings()
		finalStrs := NewVMState(vm.Poweroff, vm.Undeployed).ToStrings()
		stateConf := NewVMUpdateStateConf(timeout, transientStrs, finalStrs)

		_, err = waitForVMStates(ctx, vmc, stateConf)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to wait virtual machine to be in %s state", strings.Join(finalStrs, ",")),
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
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
			stateConf := NewVMUpdateStateConf(timeout,
				NewVMLCMState(vm.BootPoweroff).ToStrings(),
				NewVMLCMState(vm.Running).ToStrings(),
			)

			// wait for the VM to be back in RUNNING state
			_, err = waitForVMStates(ctx, vmc, stateConf)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to wait virtual machine to be in RUNNING state",
					Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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

		// wait for the VM to be RUNNING to avoid action failures
		// RUNNING state is added to transient one in case of slow cloud
		transientStrs := NewVMLCMState(vm.Running).
			Append(vmDiskTransientStates).
			Append(vmNICTransientStates).ToStrings()
		finalStrs := NewVMLCMState(vm.Running).ToStrings()
		stateConf := NewVMUpdateStateConf(timeout,
			transientStrs,
			finalStrs,
		)

		_, err = waitForVMStates(ctx, vmc, stateConf)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to wait virtual machine to be in RUNNING state",
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		log.Printf("[INFO] Update VM configuration: %s", tpl.String())

		err := vmc.UpdateConf(tpl.String())
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update VM configuration",
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		// wait for the VM to be RUNNING after update
		stateConf = NewVMUpdateStateConf(timeout,
			[]string{},
			finalStrs,
		)
		_, err = waitForVMStates(ctx, vmc, stateConf)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to wait virtual machine to be in RUNNING state",
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = vmc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
		"dev_prefix",
		"cache",
		"discard",
		"io",
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
				disk["driver"] == newDisk["driver"] &&
				disk["dev_prefix"] == newDisk["dev_prefix"] &&
				disk["cache"] == newDisk["cache"] &&
				disk["discard"] == newDisk["discard"] &&
				disk["io"] == newDisk["io"] {

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
				(len(diskConfig["dev_prefix"].(string)) > 0 && diskConfig["dev_prefix"] != cfg["computed_dev_prefix"]) ||
				(len(diskConfig["cache"].(string)) > 0 && diskConfig["cache"] != cfg["computed_cache"]) ||
				(len(diskConfig["discard"].(string)) > 0 && diskConfig["discard"] != cfg["computed_discard"]) ||
				(len(diskConfig["io"].(string)) > 0 && diskConfig["io"] != cfg["computed_io"]) ||
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

func getNICUpdates(d *schema.ResourceData) (NICUpdates, error) {

	// get unique elements of each list of configs
	// NOTE: diffListConfig relies on Set, so we may loose list ordering of NICs here
	// it's why we reorder the attach list below
	beforeNicsCfg, afterNicsCfg := d.GetChange("nic")
	beforeNicsList, ok := beforeNicsCfg.([]any)
	if !ok {
		return NICUpdates{}, fmt.Errorf("invalid old NICs configuration: %v", beforeNicsCfg)
	}
	afterNicsList, ok := afterNicsCfg.([]any)
	if !ok {
		return NICUpdates{}, fmt.Errorf("invalid updated NICs configuration: %v", afterNicsCfg)
	}
	toDetach, toAttach := getNICUpdateDiff(afterNicsList, beforeNicsList)

	newNics, err := MapArrayDifferenceByKeyValue(toAttach, toDetach, "nic_id")
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to get new NICs: %w", err)
	}

	//gets nics to be detached permanently
	nicsToBeDetachedPermanently, err := MapArrayDifferenceByKeyValue(toDetach, toAttach, "nic_id")
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to get NICs to be detached permanently: %w", err)
	}

	//get nics that are going to be reattached because they were updated
	updatedNicsToBeReattached, err := MapArrayIntersectionByKeyValue(toAttach, toDetach, "nic_id")
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to get NICs to be reattached: %w", err)
	}

	allNicsToDetach := toDetach
	allNicsToAttach := toAttach

	// in case of NICs updated in the middle of the NIC list
	// they would be reattached at the end of the list (we don't have in place XML-RPC update method).
	// keep_nic_order prevent this behavior adding more NICs to detach/attach to keep initial orderin
	toRecreateOrdered := []any{}
	if keepNicOrder, ok := d.Get("keep_nic_order").(bool); ok && keepNicOrder && len(toDetach) > 0 {
		var err error
		toRecreateOrdered, err = getNicsToRecreateOrderedList(beforeNicsList, toDetach)
		if err != nil {
			return NICUpdates{}, fmt.Errorf("failed to get NICs to recreate ordered list: %w", err)
		}

		allNicsToDetach = append(allNicsToDetach, toRecreateOrdered...)
		allNicsToAttach = append(allNicsToAttach, toRecreateOrdered...)
	}

	// reorder toAttach NIC list according to new nics list order
	orderedAllNicsToAttach, err := orderMatchingNicsByReference(afterNicsList, allNicsToAttach, nicMatchByAttributes)
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to order NICs to attach: %w", err)
	}

	return NICUpdates{
		newNics:                 newNics,
		deletedNics:             nicsToBeDetachedPermanently,
		nicsReattachedUpdated:   updatedNicsToBeReattached,
		nicsReattachedReordered: toRecreateOrdered,
		allNicsToAttach:         orderedAllNicsToAttach,
		allNicsToDetach:         allNicsToDetach,
	}, nil

}

func getNICAliasUpdates(d *schema.ResourceData, nicUpdates NICUpdates) (NICUpdates, error) {

	// get unique elements of each list of configs
	// NOTE: diffListConfig relies on Set, so we may loose list ordering of NICs here
	// it's why we reorder the attach list below
	beforeNicAliasCfg, afterNicAliasCfg := d.GetChange("nic_alias")
	beforeNicAliasList, ok := beforeNicAliasCfg.([]any)
	if !ok {
		return NICUpdates{}, fmt.Errorf("invalid old NIC Alias configuration: %v", beforeNicAliasCfg)
	}
	afterNicAliasesList, ok := afterNicAliasCfg.([]any)
	if !ok {
		return NICUpdates{}, fmt.Errorf("invalid updated NIC Alias configuration: %v", afterNicAliasCfg)
	}
	toDetach, toAttach := getNICAliasUpdateDiff(afterNicAliasesList, beforeNicAliasList)

	newNicAliases, err := MapArrayDifferenceByKeyValue(toAttach, toDetach, "nic_id")
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to get new NIC Aliases: %w", err)
	}

	//gets nics to be detached permanently
	nicAliasesToBeDetachedPermanently, err := MapArrayDifferenceByKeyValue(toDetach, toAttach, "nic_id")
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to get NIC Aliases to be detached permanently: %w", err)
	}

	//get nics that are going to be reattached because they were updated
	updatedNicAliasesToBeReattached, err := MapArrayIntersectionByKeyValue(toAttach, toDetach, "nic_id")
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to get NIC Aliases to be reattached: %w", err)
	}

	//Check which NICs are going to be recreated (because of the updates and the reordering)
	// and recreate their NICAliases as well (ignore the nic aliases that are going to be deleted)
	dependantNicAliasToBeReattached, err := retrieveDependantNICAliasesToRecreate(beforeNicAliasList, nicUpdates)
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to retrieve dependant NIC Aliases from recreated NICs: %s", err)
	}
	//remove nic aliases that are going to be detached permanently
	dependantNicAliasToBeReattached, err = MapArrayDifferenceByKeyValue(dependantNicAliasToBeReattached, nicAliasesToBeDetachedPermanently, "nic_id")
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to remove NIC Aliases that are going to be detached permanently: %w", err)
	}
	//remove the nic aliases that are going to be recreated
	dependantNicAliasToBeReattached, err = MapArrayDifferenceByKeyValue(dependantNicAliasToBeReattached, updatedNicAliasesToBeReattached, "nic_id")
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to remove NIC Aliases that are going to be reattached: %w", err)
	}
	toDetach = append(toDetach, dependantNicAliasToBeReattached...)
	toAttach = append(toAttach, dependantNicAliasToBeReattached...)

	allNicAliasesToDetach := toDetach
	allNicAliasesToAttach := toAttach

	// in case of NICs updated in the middle of the NIC list
	// they would be reattached at the end of the list (we don't have in place XML-RPC update method).
	// keep_nic_order prevent this behavior adding more NICs to detach/attach to keep initial ordering
	toRecreateOrdered := []any{}
	if keepNicOrder, ok := d.Get("keep_nic_order").(bool); ok && keepNicOrder && len(toDetach) > 0 {
		var err error
		toRecreateOrdered, err = getNicsToRecreateOrderedList(beforeNicAliasList, toDetach)
		if err != nil {
			return NICUpdates{}, fmt.Errorf("failed to get NICs to recreate ordered list: %w", err)
		}
		allNicAliasesToDetach = append(allNicAliasesToDetach, toRecreateOrdered...)
		allNicAliasesToAttach = append(allNicAliasesToAttach, toRecreateOrdered...)
	}

	// reorder nic_alias to attach list according to new nic_alias list order
	orderedAllNicAliasesToAttach, err := orderMatchingNicsByReference(afterNicAliasesList, allNicAliasesToAttach, nicAliasMatchByAttributes)
	if err != nil {
		return NICUpdates{}, fmt.Errorf("failed to order NIC aliases to attach: %w", err)
	}

	return NICUpdates{
		newNics:                 newNicAliases,
		deletedNics:             nicAliasesToBeDetachedPermanently,
		nicsReattachedUpdated:   updatedNicAliasesToBeReattached,
		nicsReattachedReordered: toRecreateOrdered,
		allNicsToAttach:         orderedAllNicAliasesToAttach,
		allNicsToDetach:         allNicAliasesToDetach,
	}, nil

}

func detachNICAliases(ctx context.Context, vmc *goca.VMController, timeout time.Duration, aliasUpdates NICUpdates) error {
	nicAliasesToDetach := aliasUpdates.allNicsToDetach
	if len(nicAliasesToDetach) == 0 {
		log.Printf("[DEBUG] No NIC Aliases to detach")
		return nil
	}

	log.Printf("[DEBUG] Detaching NIC Aliases: %v", nicAliasesToDetach)
	return detachNicAliasList(ctx, vmc, nicAliasesToDetach, timeout)
}

func attachNICAliases(ctx context.Context, vmc *goca.VMController, timeout time.Duration, aliasUpdates NICUpdates) error {
	nicAliasesToAttach := aliasUpdates.allNicsToAttach
	if len(nicAliasesToAttach) == 0 {
		log.Printf("[DEBUG] No NIC Aliases to attach")
		return nil
	}

	log.Printf("[DEBUG] Attaching NIC Aliases: %v", nicAliasesToAttach)
	return attachNicAliasList(ctx, vmc, nicAliasesToAttach, timeout)
}

func detachNICs(ctx context.Context, vmc *goca.VMController, timeout time.Duration, nicUpdates NICUpdates, aliasUpdates NICUpdates) error {
	nicsToDetach := nicUpdates.allNicsToDetach
	if len(nicsToDetach) == 0 {
		log.Printf("[DEBUG] No NICs to detach")
		return nil
	}

	if err := checkDependantNICAliases(nicUpdates, aliasUpdates); err != nil {
		return fmt.Errorf("failed dependant NIC Aliases check: %w", err)
	}

	log.Printf("[DEBUG] Detaching NICs: %v", nicsToDetach)
	return detachNicList(ctx, vmc, nicsToDetach, timeout)
}

func checkDependantNICAliases(nicUpdates NICUpdates, aliasUpdates NICUpdates) error {
	//Check if the nics to be deleted are referenced by any nic_alias (or have nic_aliases)
	dependantNicAliasMap, err := getDependantNICAliasesIDsMap(nicUpdates.deletedNics)
	if err != nil {
		return fmt.Errorf("failed to get NICs dependant NIC Aliases: %w", err)
	}
	// if dependant nicAliases are not going to be deleted as well, throw an error
	nicAliasIDsToDetach, err := aliasUpdates.getNICsToDetachIds()
	if err != nil {
		return fmt.Errorf("failed to get NIC Aliases to detach IDs: %w", err)
	}

	for nicId, nicAliasIds := range dependantNicAliasMap {
		if len(nicAliasIds) == 0 {
			// no dependant nic_aliases for this nic
			continue
		}
		// get elements that are in the dependant nicAliasIds but not in nicAliasIDsToDetach
		nicAliasToSurvive := ArrayDifference(nicAliasIds, nicAliasIDsToDetach)
		// if there are nic_aliases that are not going to be deleted, throw an error
		if len(nicAliasToSurvive) > 0 {
			return fmt.Errorf("referential error: NIC with ID %d referenced by NIC Aliases that are not going to be detached: %v", nicId, nicAliasToSurvive)
		}
	}
	return nil
}

func attachNICs(ctx context.Context, vmc *goca.VMController, timeout time.Duration, nicUpdates NICUpdates) error {
	nicsToAttach := nicUpdates.allNicsToAttach
	if len(nicsToAttach) == 0 {
		log.Printf("[DEBUG] No NICs to attach")
		return nil
	}

	log.Printf("[DEBUG] Attaching NICs: %v", nicsToAttach)
	return attachNicList(ctx, vmc, nicsToAttach, timeout)
}

// returns a map of NIC IDs to their aliases NIC IDs
func getDependantNICAliasesIDsMap(nics []any) (map[int][]int, error) {
	nicAliasesMap := make(map[int][]int)
	for _, nic := range nics {
		nicCfg, ok := nic.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid NIC configuration: %v", nic)
		}
		nicID, ok := nicCfg["nic_id"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid nic_id in NIC configuration: %v", nicCfg)
		}
		nicAliasesIDs, err := parseNICAliasIds(nicCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse NIC Aliases IDs for NIC ID %d: %w", nicID, err)
		}
		nicAliasesMap[nicID] = nicAliasesIDs
	}
	return nicAliasesMap, nil
}

func parseNICAliasIds(nic map[string]any) ([]int, error) {
	nicAliasesIdsList := []int{}
	nicAliasesStr, ok := nic["computed_alias_ids"].(string)
	if ok && len(nicAliasesStr) > 0 {
		nicAliases := strings.Split(nicAliasesStr, ",")
		for _, nicAlias := range nicAliases {
			nicAlias = strings.TrimSpace(nicAlias)
			if len(nicAlias) == 0 {
				continue
			}
			nicAliasId, err := strconv.Atoi(nicAlias)
			if err != nil {
				return nil, fmt.Errorf("invalid NIC Alias ID: %s in NIC configuration: %v", nicAlias, nic)
			}
			nicAliasesIdsList = append(nicAliasesIdsList, nicAliasId)
		}
	}
	return nicAliasesIdsList, nil
}

func getNICUpdateDiff(oldNics []any, updatedNics []any) ([]any, []any) {
	return diffListConfig(oldNics, updatedNics,
		&schema.Resource{
			Schema: nicFields(),
		},
		"name",
		"network_id",
		"ip",
		"ip6",
		"ip6_ula",
		"ip6_global",
		"ip6_link",
		"mac",
		"security_groups",
		"model",
		"virtio_queues",
		"physical_device",
		"method",
		"gateway",
		"dns",
		"network_mode_auto",
		"sched_requirements",
		"sched_rank",
	)
}

func getNICAliasUpdateDiff(oldNicAliases []any, updatedNicAliases []any) ([]any, []any) {
	return diffListConfig(oldNicAliases, updatedNicAliases,
		&schema.Resource{
			Schema: nicAliasFields(),
		},
		"name",
		"network",
		"network_id",
		"parent",
		"ip",
		"ip6",
		"ip6_ula",
		"ip6_global",
		"ip6_link",
		"mac",
		"security_groups",
		"gateway",
		"dns",
	)
}

func getNicsToRecreateOrderedList(oldNicsList []any, toDetach []any) ([]any, error) {
	detachedNicsMap, detachedNicIDs, err := getNicsMap(toDetach)
	if err != nil {
		return nil, fmt.Errorf("failed to generate detached NICs map: %w", err)
	}
	minNICId := detachedNicIDs[0]

	oldNicsMap, oldNicsIDs, err := getNicsMap(oldNicsList)
	if err != nil {
		return nil, fmt.Errorf("failed to generate old NICs map: %w", err)
	}

	// get the indexes greater than or equal to minNIC
	idx := sort.SearchInts(oldNicsIDs, minNICId)
	if idx < 0 || idx >= len(oldNicsIDs) {
		return nil, fmt.Errorf("failed to find the minimum NIC ID %d in old NICs list", minNICId)
	}

	// nics that have IDs greater than or equial to minNIC should be recreated
	nicsToRecreate := make([]any, 0, len(oldNicsIDs)-idx)
	for _, oldNicID := range oldNicsIDs[idx:] {
		if _, ok := detachedNicsMap[oldNicID]; ok {
			continue // skip NICs that are already in toDetach
		}
		nicsToRecreate = append(nicsToRecreate, oldNicsMap[oldNicID])
	}
	return nicsToRecreate, nil
}

// generates a map of NICs with their IDs as keys and a sorted slice of NIC IDs
func getNicsMap(nicsList []any) (map[int]any, []int, error) {
	nicsMap := make(map[int]any, len(nicsList))
	sortedNicIDs := make([]int, 0, len(nicsList))
	for _, nic := range nicsList {
		nicMap, ok := nic.(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("invalid NIC configuration: %v", nic)
		}
		nicID, ok := nicMap["nic_id"].(int)
		if !ok {
			return nil, nil, fmt.Errorf("invalid nic_id in NIC configuration: %v", nicMap)
		}
		nicsMap[nicID] = nicMap
		sortedNicIDs = append(sortedNicIDs, nicID)
	}
	sort.Ints(sortedNicIDs)
	return nicsMap, sortedNicIDs, nil
}

func orderMatchingNicsByReference(nicsReferenceList []any, nicsToOrderList []any, matchingFunction func(map[string]any, map[string]any) (bool, error)) ([]any, error) {

	orderedNics := make([]any, len(nicsToOrderList))
	i := 0
	for _, nicRef := range nicsReferenceList {
		nicRefCfg, ok := nicRef.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid NIC reference configuration: %v", nicRef)
		}
		mathcingNic, found, err := findMatchingNic(nicRefCfg, nicsToOrderList, matchingFunction)
		if err != nil {
			return nil, fmt.Errorf("failed to find matching NIC for reference: %v, error: %w", nicRefCfg, err)
		}
		if !found {
			continue
		}
		orderedNics[i] = mathcingNic
		i++
		// no more nics to order
		if i >= len(nicsToOrderList) {
			break
		}
	}
	if i < len(nicsToOrderList) {
		return nil, fmt.Errorf("not all NICs were ordered, expected: %d, got: %d", len(nicsToOrderList), i)
	}
	return orderedNics, nil
}

func findMatchingNic(referenceNic map[string]any, nicsList []any, matchingFunction func(map[string]any, map[string]any) (bool, error)) (map[string]any, bool, error) {
	for _, nic := range nicsList {
		nicMap, ok := nic.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("invalid NIC configuration: %v", nic)
		}

		match, err := matchingFunction(referenceNic, nicMap)
		if err != nil {
			return nil, false, fmt.Errorf("failed to match NICs by attributes: %w", err)
		}

		if match {
			return nicMap, true, nil
		}
	}
	return nil, false, nil
}

func nicMatchByAttributes(nic map[string]any, otherNic map[string]any) (bool, error) {

	nicSecGroup, ok := nic["security_groups"].([]any)
	if !ok {
		return false, fmt.Errorf("invalid security_groups in NIC configuration: %v", nic)
	}

	otherNICSecGroup, ok := otherNic["security_groups"].([]any)
	if !ok {
		return false, fmt.Errorf("invalid security_groups in other NIC configuration: %v", otherNic)
	}

	matches := ArrayToString(nicSecGroup, ",") == ArrayToString(otherNICSecGroup, ",") &&
		nic["ip"] == otherNic["ip"] &&
		nic["ip6"] == otherNic["ip6"] &&
		nic["ip6_ula"] == otherNic["ip6_ula"] &&
		nic["ip6_global"] == otherNic["ip6_global"] &&
		nic["ip6_link"] == otherNic["ip6_link"] &&
		nic["mac"] == otherNic["mac"] &&
		nic["model"] == otherNic["model"] &&
		nic["virtio_queues"] == otherNic["virtio_queues"] &&
		nic["physical_device"] == otherNic["physical_device"]

	return matches, nil
}

func nicAliasMatchByAttributes(nicAlias map[string]any, otherNicAlias map[string]any) (bool, error) {

	nicAliasSecGroup, ok := nicAlias["security_groups"].([]any)
	if !ok {
		return false, fmt.Errorf("invalid security_groups in NIC Alias configuration: %v", nicAlias)
	}

	otherNICAliasSecGroup, ok := otherNicAlias["security_groups"].([]any)
	if !ok {
		return false, fmt.Errorf("invalid security_groups in other NIC Alias configuration: %v", otherNicAlias)
	}

	matches := ArrayToString(nicAliasSecGroup, ",") == ArrayToString(otherNICAliasSecGroup, ",") &&
		nicAlias["ip"] == otherNicAlias["ip"] &&
		nicAlias["ip6"] == otherNicAlias["ip6"] &&
		nicAlias["ip6_ula"] == otherNicAlias["ip6_ula"] &&
		nicAlias["ip6_global"] == otherNicAlias["ip6_global"] &&
		nicAlias["ip6_link"] == otherNicAlias["ip6_link"] &&
		nicAlias["mac"] == otherNicAlias["mac"] &&
		nicAlias["gateway"] == otherNicAlias["gateway"] &&
		nicAlias["dns"] == otherNicAlias["dns"] &&
		nicAlias["network"] == otherNicAlias["network"] &&
		nicAlias["network_id"] == otherNicAlias["network_id"] &&
		nicAlias["parent"] == otherNicAlias["parent"]

	return matches, nil
}

func detachNicList(ctx context.Context, vmc *goca.VMController, nicList []any, timeout time.Duration) error {
	for _, nic := range nicList {
		nicConfig, ok := nic.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid NIC configuration: %v", nic)
		}

		nicID, ok := nicConfig["nic_id"].(int)
		if !ok {
			return fmt.Errorf("invalid nic_id in NIC configuration: %v", nicConfig)
		}

		err := vmNICDetach(ctx, vmc, timeout, nicID)
		if err != nil {
			return fmt.Errorf("vm nic detach: %s", err)

		}
	}
	return nil
}

func attachNicList(ctx context.Context, vmc *goca.VMController, nicList []any, timeout time.Duration) error {
	for _, nic := range nicList {
		nicConfig, ok := nic.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid NIC configuration: %v", nic)
		}

		nicTpl := makeNICVector(nicConfig)

		_, err := vmNICAttach(ctx, vmc, timeout, nicTpl)
		if err != nil {
			return fmt.Errorf("vm nic attach: %s", err)
		}
	}
	return nil
}

func detachNicAliasList(ctx context.Context, vmc *goca.VMController, nicAliasList []any, timeout time.Duration) error {
	for _, nicAlias := range nicAliasList {
		nicConfig, ok := nicAlias.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid NIC Alias configuration: %v", nicAlias)
		}

		nicID, ok := nicConfig["nic_id"].(int)
		if !ok {
			return fmt.Errorf("invalid nic_id in NIC Alias configuration: %v", nicConfig)
		}

		err := vmNICAliasDetach(ctx, vmc, timeout, nicID)
		if err != nil {
			return fmt.Errorf("vm nic detach: %s", err)

		}
	}
	return nil
}

func attachNicAliasList(ctx context.Context, vmc *goca.VMController, nicAliasList []any, timeout time.Duration) error {
	for _, nicAlias := range nicAliasList {
		nicAliasConfig, ok := nicAlias.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid NIC Alias configuration: %v", nicAlias)
		}

		nicAliasTpl := makeNICAliasVector(nicAliasConfig)

		_, err := vmNICAliasAttach(ctx, vmc, timeout, nicAliasTpl)
		if err != nil {
			return fmt.Errorf("vm nic alias attach: %s", err)
		}
	}
	return nil
}

// ctyStringMapToGoMap converts a cty.Value string map to a Go map.
func ctyStringMapToGoMap(val cty.Value) map[string]interface{} {
	result := make(map[string]interface{})
	if val.IsNull() || !val.IsKnown() {
		return result
	}
	val.ForEachElement(func(key cty.Value, val cty.Value) (stop bool) {
		if key.Type() == cty.String && val.Type() == cty.String {
			result[key.AsString()] = val.AsString()
		}
		return false
	})
	return result
}

// getContextWo retrieves write-only context variables from the raw Terraform config.
// WriteOnly attributes are never in state, so they must be read via GetRawConfigAt.
func getContextWo(d *schema.ResourceData) map[string]interface{} {
	val, diags := d.GetRawConfigAt(cty.GetAttrPath("context_wo"))
	if diags.HasError() {
		return make(map[string]interface{})
	}
	return ctyStringMapToGoMap(val)
}

// computeContextWoHash computes a deterministic SHA256 hash of context_wo values.
// Returns empty string if context_wo is empty.
func computeContextWoHash(contextWo map[string]interface{}) string {
	if len(contextWo) == 0 {
		return ""
	}

	keys := make([]string, 0, len(contextWo))
	for k := range contextWo {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(contextWo[k].(string))
		sb.WriteString(";")
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// updateVMTemplateVec update a vector of an existing VM template
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

	// Add new elements (or update elements not tracked in appliedCfg, like context_wo)
	for key, value := range newCfg {
		keyUp := strings.ToUpper(key)

		_, ok := appliedCfg[keyUp]
		if ok {
			continue
		}

		// Delete first to ensure we're not creating duplicates
		// This is important for context_wo values which aren't in appliedCfg
		targetVec.Del(keyUp)
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
			Summary:  "Failed to get the virtual machine controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// wait state to be ready
	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultVMTimeout {
		timeout = d.Timeout(schema.TimeoutDelete)
	}

	// wait for the VM to be in a state that permit it to be deleted
	finalStrs := vmDeleteReadyStates.ToStrings()
	stateConf := NewVMUpdateStateConf(timeout,
		[]string{},
		finalStrs,
	)

	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Failed to wait virtual machine to be in %s state", strings.Join(finalStrs, ",")),
			Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
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
			Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// wait for the VM to be powered off
	// RUNNING state is added to transient one in case of slow cloud
	transientStrs := NewVMLCMState(vm.Running).
		Append(vmDeleteTransientStates).
		ToStrings()

	stateConf = NewVMStateConf(timeout,
		transientStrs,
		NewVMState(vm.Done).ToStrings(),
	)

	ret, err := waitForVMStates(ctx, vmc, stateConf)
	if err != nil {

		log.Printf("[WARN] waitForVMStates: %s\n", err)

		// retry
		if ret != nil {

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
						Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
					})
					return diags
				}

				_, err = waitForVMStates(ctx, vmc, stateConf)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to wait virtual machine to be in DONE state",
						Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
					})
					return diags
				}

			} else {
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to wait virtual machine to be in DONE state",
						Detail:   fmt.Sprintf("virtual machine (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
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

	log.Printf("[INFO] Successfully terminated VM\n")
	return nil
}

func generateVm(d *schema.ResourceData, meta interface{}, templateContent *vm.Template) (*vm.Template, error) {

	tpl := vm.NewTemplate()

	if d.Get("name") != nil {
		tpl.Add(vmk.Name, d.Get("name").(string))
	}

	// Generate CONTEXT by merging regular context with write-only context_wo.
	context := d.Get("context").(map[string]interface{})
	contextWo := getContextWo(d)
	log.Printf("Number of CONTEXT vars: %d, CONTEXT_WO vars: %d", len(context), len(contextWo))

	var tplContext *dyn.Vector
	if templateContent != nil {
		tplContext, _ = templateContent.GetVector(vmk.ContextVec)
	}

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

		// Add write-only context variables
		for key, value := range contextWo {
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

		// Add write-only context elements to the template
		for key, value := range contextWo {
			keyUp := strings.ToUpper(key)
			tpl.AddCtx(vmk.Context(keyUp), fmt.Sprint(value))
		}
	}

	err := generateVMTemplate(d, tpl)
	if err != nil {
		return tpl, err
	}

	// add default tags if they aren't overriden
	config := meta.(*Configuration)

	if len(config.defaultTags) > 0 {
		for k, v := range config.defaultTags {
			key := strings.ToUpper(k)

			p, _ := tpl.GetPair(key)
			if p != nil {
				continue
			}

			// keep the tag from the template
			if templateContent != nil {
				p, _ := templateContent.GetPair(key)
				if p != nil {
					continue
				}
			}

			tpl.AddPair(key, v)
		}
	}

	return tpl, nil
}

func addNICs(d *schema.ResourceData, tpl *vm.Template) {
	//Generate NIC definition
	nics := d.Get("nic").([]interface{})
	log.Printf("Number of NICs: %d", len(nics))

	for i := 0; i < len(nics); i++ {
		nicconfig := nics[i].(map[string]interface{})

		nic := makeNICVector(nicconfig)
		tpl.Elements = append(tpl.Elements, nic)
	}
}

func addNICAliases(d *schema.ResourceData, tpl *vm.Template) {
	nicAliases := d.Get("nic_alias").([]interface{})
	log.Printf("Number of NIC Aliases: %d", len(nicAliases))

	for i := 0; i < len(nicAliases); i++ {
		nicAliasConfig := nicAliases[i].(map[string]interface{})

		nicAliasVector := makeNICAliasVector(nicAliasConfig)
		tpl.Elements = append(tpl.Elements, nicAliasVector)
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
				diff.ForceNew(fmt.Sprintf("disk.%d.dev_prefix", i))
				diff.ForceNew(fmt.Sprintf("disk.%d.cache", i))
				diff.ForceNew(fmt.Sprintf("disk.%d.discard", i))
				diff.ForceNew(fmt.Sprintf("disk.%d.io", i))
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

	SetVMTagsDiff(ctx, diff, v)

	// Detect context_wo changes via hash comparison.
	// WriteOnly attributes don't trigger updates, so we update context_wo_hash
	// via SetNew to create a diff that triggers UPDATE.
	// Detect context_wo changes via hash comparison.
	// WriteOnly attributes don't trigger updates on their own, so we update
	// context_wo_hash via SetNew to create a diff that triggers UPDATE.
	if diff.Id() != "" {
		contextWo := ctyStringMapToGoMap(diff.GetRawPlan().GetAttr("context_wo"))
		currentHash := computeContextWoHash(contextWo)
		previousHash := diff.Get("context_wo_hash").(string)

		if currentHash != previousHash && currentHash != "" {
			if err := diff.SetNew("context_wo_hash", currentHash); err != nil {
				return fmt.Errorf("failed to update context_wo_hash: %w", err)
			}
		}
	}

	return nil
}
