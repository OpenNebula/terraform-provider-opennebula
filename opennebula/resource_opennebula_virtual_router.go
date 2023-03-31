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
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	vr "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualrouter"
	vrk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualrouter/keys"
)

func resourceOpennebulaVirtualRouter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualRouterCreate,
		ReadContext:   resourceOpennebulaVirtualRouterRead,
		Exists:        resourceOpennebulaVirtualRouterExists,
		UpdateContext: resourceOpennebulaVirtualRouterUpdate,
		DeleteContext: resourceOpennebulaVirtualRouterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the virtual router",
			},
			"instance_template_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the template of the virtual router instances.",
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the virtual router (in Unix format, owner-group-other, use-manage-admin)",
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
				Description: "ID of the user that will own the virtual router",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the virtual router",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the virtual router",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the virtual router",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the virtual router, If empty, it uses caller group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the entity",
			},
			"lock":             lockSchema(),
			"tags":             tagsSchema(),
			"default_tags":     defaultTagsSchemaComputed(),
			"tags_all":         tagsSchemaComputed(),
			"template_section": templateSectionSchema(),
		},
	}
}

func getVirtualRouterController(d *schema.ResourceData, meta interface{}) (*goca.VirtualRouterController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	vrID, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.VirtualRouter(int(vrID)), nil
}

func changeVirtualRouterGroup(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller
	var gid int

	vrc, err := getVirtualRouterController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		gid, err = controller.Groups().ByName(d.Get("group").(string))
		if err != nil {
			return err
		}
	}

	err = vrc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaVirtualRouterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller

	vrDef := generateVirtualRouter(d, meta)

	vrID, err := controller.VirtualRouters().Create(vrDef)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router creation failed",
			Detail:   err.Error(),
		})
		return diags
	}

	vrc := controller.VirtualRouter(vrID)

	d.SetId(fmt.Sprintf("%v", vrID))

	// Change Permissions only if Permissions are set
	if perms, ok := d.GetOk("permissions"); ok {
		err = vrc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router permission change failed",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err = changeVirtualRouterGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router group change failed",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
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
				Summary:  "virtual router wrong lock level",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = vrc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router group lock failed",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaVirtualRouterRead(ctx, d, meta)
}

func resourceOpennebulaVirtualRouterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get requested virtual router from all virtual routers
	vrc, err := getVirtualRouterController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router error",
			Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	vr, err := vrc.Info(false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual router %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router info error",
			Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", vr.ID))
	d.Set("name", vr.Name)
	d.Set("uid", vr.UID)
	d.Set("gid", vr.GID)
	d.Set("uname", vr.UName)
	d.Set("gname", vr.GName)
	d.Set("permissions", permissionsUnixString(*vr.Permissions))

	desc, _ := vr.Template.Get("DESCRIPTION")
	d.Set("description", desc)

	if vr.LockInfos != nil {
		d.Set("lock", LockLevelToString(vr.LockInfos.Locked))
	}

	config := meta.(*Configuration)

	err = flattenTemplateSection(d, meta, &vr.Template.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten template section",
			Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	tags := make(map[string]interface{})
	tagsAll := make(map[string]interface{})

	vrTpl := vr.Template
	for i, _ := range vrTpl.Elements {
		pair, ok := vrTpl.Elements[i].(*dyn.Pair)
		if !ok {
			continue
		}

		switch pair.Key() {
		case "DESCRIPTION":
			desc, err := vrTpl.Get("DESCRIPTION")
			if desc != "" {
				err = d.Set("description", desc)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "virtual router set attribute error",
						Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
		default:
		}
	}

	// Get default tags
	oldDefault := d.Get("default_tags").(map[string]interface{})
	for k, _ := range oldDefault {
		key := strings.ToUpper(k)
		tagValue, err := vrTpl.GetStr(key)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to get default tag",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
		}
		tagsAll[k] = tagValue
	}
	d.Set("default_tags", config.defaultTags)

	// Get only tags described in the configuration
	if tagsInterface, ok := d.GetOk("tags"); ok {
		for k, _ := range tagsInterface.(map[string]interface{}) {
			tagValue, err := vrTpl.GetStr(strings.ToUpper(k))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Failed to get tag from the template",
					Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
				})
			}
			tags[k] = tagValue
			tagsAll[k] = tagValue
		}

		err := d.Set("tags", tags)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router set attribute error",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
		}
	}
	d.Set("tags_all", tagsAll)

	return diags
}

func resourceOpennebulaVirtualRouterExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	vRouterID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.VirtualRouter(int(vRouterID)).Info(false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaVirtualRouterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	//Get virtual router
	vrc, err := getVirtualRouterController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router error",
			Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	vrInfos, err := vrc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router info error",
			Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	lock, lockOk := d.GetOk("lock")
	if d.HasChange("lock") && lockOk && lock.(string) == "UNLOCK" {

		err = vrc.Unlock()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router unlock error",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("name") {
		err := vrc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router rename error",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		// Update Name in internal struct
		vrInfos, err = vrc.Info(false)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router info error",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name for tpl %s\n", vrInfos.Name)
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = vrc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "virtual router permission change failed",
					Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated virtual router %s\n", vrInfos.Name)
	}

	if d.HasChange("group") {
		err = changeVirtualRouterGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router group change failed",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated group for virtual router %s\n", vrInfos.Name)
	}

	update := false
	newTpl := vr.NewTemplate()
	if d.HasChange("description") {
		update = true
		newTpl.Add("DESCRIPTION", d.Get("description").(string))
	}

	if d.HasChange("template_section") {

		updateTemplateSection(d, &newTpl.Template)

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
			newTpl.Del(strings.ToUpper(k))
		}

		// add/update tags
		for k, v := range newTags {
			newTpl.Del(strings.ToUpper(k))
			newTpl.AddPair(strings.ToUpper(k), v)
		}

		update = true
	}

	if update {
		err = vrc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router update failed",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
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
				Summary:  "virtual router wrong lock level",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = vrc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual router group lock failed",
				Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaVirtualRouterRead(ctx, d, meta)
}

func resourceOpennebulaVirtualRouterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	vrc, err := getVirtualRouterController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router error",
			Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	err = vrc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router delete failed",
			Detail:   fmt.Sprintf("virtual router (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully deleted virtual router ID %s\n", d.Id())

	return nil
}

func generateVirtualRouter(d *schema.ResourceData, meta interface{}) string {
	name := d.Get("name").(string)

	tpl := vr.NewTemplate()

	tpl.Add(vrk.Name, name)

	templateID := d.Get("instance_template_id")
	tpl.Add("TEMPLATE_ID", templateID)

	descr, ok := d.GetOk("description")
	if ok {
		tpl.Add("DESCRIPTION", descr.(string))
	}

	vectorsInterface := d.Get("template_section").(*schema.Set).List()
	if len(vectorsInterface) > 0 {
		addTemplateVectors(vectorsInterface, &tpl.Template)
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
	log.Printf("[INFO] Template definitions: %s", tplStr)

	return tplStr
}
