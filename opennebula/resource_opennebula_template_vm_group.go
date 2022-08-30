package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vmgroup"
)

func resourceOpennebulaVMGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVMGroupCreate,
		ReadContext:   resourceOpennebulaVMGroupRead,
		Exists:        resourceOpennebulaVMGroupExists,
		UpdateContext: resourceOpennebulaVMGroupUpdate,
		DeleteContext: resourceOpennebulaVMGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the template",
			},
			"role": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "Roles of the VM Group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"host_affined": {
							Type:        schema.TypeList,
							Computed:    true,
							Optional:    true,
							Description: "Host(s) where the VM should run",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"host_anti_affined": {
							Type:        schema.TypeList,
							Computed:    true,
							Optional:    true,
							Description: "Host(s) where the VM should not run",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"policy": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								validpolicies := []string{"NONE", "AFFINED", "ANTI_AFFINED"}
								value := v.(string)

								if inArray(value, validpolicies) < 0 {
									errors = append(errors, fmt.Errorf("Policy value %q must be one of: %s", k, strings.Join(validpolicies, ",")))
								}

								return
							},
						},
					},
				},
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the template vm group (in Unix format, owner-group-other, use-manage-admin)",
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
				Description: "ID of the user that will own the template vm group",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the template vm group",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the template vm group",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the template vm group",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the Template VM Group, If empty, it uses caller group",
			},
			"tags": tagsSchema(),
		},
	}
}

func getVMGroupController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.VMGroupController, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	var vmgc *goca.VMGroupController

	// Try to find the vm group by ID, if specified
	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 0)
		if err != nil {
			return nil, err
		}
		vmgc = controller.VMGroup(int(gid))
	}

	// Otherwise, try to find the template by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.VMGroups().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
		}
		vmgc = controller.VMGroup(gid)
	}

	return vmgc, nil
}

func changeVMGroupGroup(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller
	var gid int

	vmgc, err := getVMGroupController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		group := d.Get("group").(string)
		gid, err = controller.Groups().ByName(group)
		if err != nil {
			return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
		}
	}

	err = vmgc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaVMGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	vmg, err := generateVMGroup(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate description",
			Detail:   err.Error(),
		})
		return diags
	}
	vmgID, err := controller.VMGroups().Create(vmg)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the VM group",
			Detail:   err.Error(),
		})
		return diags
	}

	vmgc := controller.VMGroup(vmgID)

	d.SetId(fmt.Sprintf("%v", vmgID))

	// Change Permissions only if Permissions are set
	if perms, ok := d.GetOk("permissions"); ok {
		err = vmgc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",
				Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("group") != "" {
		err = changeVMGroupGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaVMGroupRead(ctx, d, meta)
}

func resourceOpennebulaVMGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	// Get requested template from all templates
	vmgc, err := getVMGroupController(d, meta, -2, -1, -1)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual machine group template %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the VM group controller",
			Detail:   err.Error(),
		})
		return diags

	}

	vmg, err := vmgc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", vmg.ID))
	d.Set("name", vmg.Name)
	d.Set("uid", vmg.UID)
	d.Set("gid", vmg.GID)
	d.Set("uname", vmg.UName)
	d.Set("gname", vmg.GName)
	d.Set("permissions", permissionsUnixString(*vmg.Permissions))

	// Get Human readable vmg information
	err = flattenVMGroupRoles(d, vmg.Roles)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten roles",
			Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	err = flattenVMGroupTags(d, &vmg.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten tags",
			Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return nil
}

func flattenVMGroupTags(d *schema.ResourceData, t *dyn.Template) error {

	tags := make(map[string]interface{})
	var err error
	// Get only tags from userTemplate
	tagsInterface, ok := d.GetOk("tags")
	if ok {
		for k, _ := range tagsInterface.(map[string]interface{}) {
			tags[k], err = t.GetStr(strings.ToUpper(k))
			if err != nil {
				return err
			}
		}
	}

	if ok {
		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func flattenVMGroupRoles(d *schema.ResourceData, vmgRoles []vmgroup.Role) error {
	var roles []map[string]interface{}

	for _, vmgr := range vmgRoles {

		hAff := make([]int, 0)
		hAntiAff := make([]int, 0)
		if vmgr.HostAffined != "" {
			hostAffString := strings.Split(vmgr.HostAffined, ",")
			for _, h := range hostAffString {
				hostAffInt, _ := strconv.ParseInt(h, 10, 0)
				hAff = append(hAff, int(hostAffInt))
			}
		}
		if vmgr.HostAntiAffined != "" {
			hostAntiAffString := strings.Split(vmgr.HostAntiAffined, ",")
			for _, h := range hostAntiAffString {
				hostAntiAffInt, _ := strconv.ParseInt(h, 10, 0)
				hAntiAff = append(hAff, int(hostAntiAffInt))
			}
		}
		roles = append(roles, map[string]interface{}{
			"id":                vmgr.ID,
			"name":              vmgr.Name,
			"host_affined":      hAff,
			"host_anti_affined": hAntiAff,
			"policy":            vmgr.Policy,
		})
	}

	return d.Set("role", roles)
}

func resourceOpennebulaVMGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	serviceTemplateID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.VMGroup(int(serviceTemplateID)).Info(false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaVMGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	//Get VMGroup
	vmgc, err := getVMGroupController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the VM group controller",
			Detail:   err.Error(),
		})
		return diags
	}
	vmg, err := vmgc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if d.HasChange("name") {
		err := vmgc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		vmg, err = vmgc.Info(false)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve informations",
				Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name for vmg %s\n", vmg.Name)
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = vmgc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change permissions",
					Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated VMGroup %s\n", vmg.Name)
	}

	if d.HasChange("group") {
		err = changeVMGroupGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated group for VMGroup %s\n", vmg.Name)
	}

	if d.HasChange("tags") {
		tagsInterface := d.Get("tags").(map[string]interface{})
		for k, v := range tagsInterface {
			vmg.Template.Del(strings.ToUpper(k))
			vmg.Template.AddPair(strings.ToUpper(k), v.(string))
		}

		err = vmgc.Update(vmg.Template.String(), 1)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("role") && d.Get("role") != "" {
		var err error

		vmg, err := generateVMGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate description",
				Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = vmgc.Update(vmg, 0)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		log.Printf("[INFO] Successfully updated Virtual Machine Group %s\n", d.Id())
	}

	return resourceOpennebulaVMGroupRead(ctx, d, meta)
}

func resourceOpennebulaVMGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vmgc, err := getVMGroupController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the VM group controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = vmgc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("VM group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	log.Printf("[INFO] Successfully deleted VMGroup ID %s\n", d.Id())

	return nil
}

func generateVMGroup(d *schema.ResourceData, meta interface{}) (string, error) {

	tpl := dyn.NewTemplate()

	tpl.AddPair("NAME", d.Get("name").(string))

	// Add Roles to the template
	roles := d.Get("role").([]interface{})
	for _, r := range roles {

		role := r.(map[string]interface{})
		rHostAff := ArrayToString(role["host_affined"].([]interface{}), ",")
		rHostAntiAff := ArrayToString(role["host_anti_affined"].([]interface{}), ",")

		roleTpl := tpl.AddVector("ROLE")
		roleTpl.AddPair("NAME", role["name"].(string))
		roleTpl.AddPair("HOST_AFFINED", rHostAff)
		roleTpl.AddPair("HOST_ANTI_AFFINED", rHostAntiAff)
		roleTpl.AddPair("POLICY", role["policy"].(string))

	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
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
			tpl.AddPair(key, v)
		}
	}

	tplStr := tpl.String()
	log.Printf("[INFO] VM group definition: %s", tplStr)

	return tplStr, nil
}
