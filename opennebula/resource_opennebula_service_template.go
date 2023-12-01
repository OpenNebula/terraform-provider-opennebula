package opennebula

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	srv_tmpl "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/service_template"
)

func resourceOpennebulaServiceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaServiceTemplateCreate,
		ReadContext:   resourceOpennebulaServiceTemplateRead,
		Exists:        resourceOpennebulaServiceTemplateExists,
		UpdateContext: resourceOpennebulaServiceTemplateUpdate,
		DeleteContext: resourceOpennebulaServiceTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the Service Template",
			},
			"template": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service Template body in json format",
				// Check JSON structure diffs, not binary diffs
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					old_template := &srv_tmpl.ServiceTemplate{}
					new_template := &srv_tmpl.ServiceTemplate{}

					err := json.Unmarshal([]byte(old), &old_template)
					if err != nil {
						return false
					}

					err = json.Unmarshal([]byte(new), &new_template)
					if err != nil {
						return false
					}

					// Custom deepEqual func to avoid empty fields #468
					return deepEqualIgnoreEmpty(old_template, new_template)
				},
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the Service Template (in Unix format, owner-group-other, use-manage-admin)",
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
				Optional:    true,
				Computed:    true,
				Description: "ID of the user that will own the Service Template",
			},
			"gid": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "ID of the group that will own the Service Template",
			},
			"uname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the user that will own the Service Template",
			},
			"gname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the group that will own the Service Template",
			},
		},
	}
}

func resourceOpennebulaServiceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	if !config.isFlowConfigured() {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Flow client isn't configured",
			Detail:   fmt.Sprintf("Check flow_endpoint in the provider configuration"),
		})
		return diags
	}

	// Marshall the json
	stemplate := &srv_tmpl.ServiceTemplate{}
	err := json.Unmarshal([]byte(d.Get("template").(string)), stemplate)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse service template json description",
			Detail:   err.Error(),
		})
		return diags
	}

	err = controller.STemplates().Create(stemplate)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the service template",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", stemplate.ID))
	stc := controller.STemplate(stemplate.ID)

	// Set the permissions on the service template if it was defined,
	// otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = stc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if _, ok := d.GetOkExists("gid"); d.Get("gname") != "" || ok {
		err = changeServiceTemplateGroup(d, meta, stc)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if _, ok := d.GetOkExists("uid"); d.Get("uname") != "" || ok {
		err = changeServiceTemplateOwner(d, meta, stc)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change owner",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("name") != "" {
		err = changeServiceTemplateName(d, meta, stc)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaServiceTemplateRead(ctx, d, meta)
}

func resourceOpennebulaServiceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)

	var diags diag.Diagnostics

	if !config.isFlowConfigured() {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Flow client isn't configured",
			Detail:   fmt.Sprintf("Check flow_endpoint in the provider configuration"),
		})
		return diags
	}

	stc, err := getServiceTemplateController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the service template controller",
			Detail:   err.Error(),
		})
		return diags
	}

	st, err := stc.Info()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", st.ID))
	d.Set("name", st.Name)
	d.Set("uid", st.UID)
	d.Set("gid", st.GID)
	d.Set("uname", st.UName)
	d.Set("gname", st.GName)
	err = d.Set("permissions", permissionsUnixString(*st.Permissions))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set attribute",
			Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// Get service.Template as map
	tmpl_byte, err := json.Marshal(st.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate json description",
			Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	d.Set("template", "{\"TEMPLATE\":"+string(tmpl_byte)+"}")

	return nil
}

func resourceOpennebulaServiceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)

	var diags diag.Diagnostics

	if !config.isFlowConfigured() {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Flow client isn't configured",
			Detail:   fmt.Sprintf("Check flow_endpoint in the provider configuration"),
		})
		return diags
	}

	//Get Service
	stc, err := getServiceTemplateController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the service template controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = stc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	log.Printf("[INFO] Successfully terminated service template\n")

	return nil
}

func resourceOpennebulaServiceTemplateExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	serviceTemplateID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.STemplate(int(serviceTemplateID)).Info()
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaServiceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	if !config.isFlowConfigured() {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Flow client isn't configured",
			Detail:   fmt.Sprintf("Check flow_endpoint in the provider configuration"),
		})
		return diags
	}

	//Get Service controller
	stc, err := getServiceTemplateController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed retrieve controller",
			Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	stemplate, err := stc.Info()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if d.HasChange("name") {
		err := stc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		stemplate, err := stc.Info()
		log.Printf("[INFO] Successfully updated name (%s) for service template ID %x\n", stemplate.Name, stemplate.ID)
	}

	if d.HasChange("permissions") && d.Get("permissions") != "" {
		if perms, ok := d.GetOk("permissions"); ok {
			err = stc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change permissions",
					Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated Permissions for service template %s\n", stemplate.Name)
	}

	if d.HasChange("gid") {
		gid := d.Get("gid").(int)
		group, err := controller.Group(gid).Info(true)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve informations",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = stc.Chgrp(gid)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.Set("gname", group.Name)
		log.Printf("[INFO] Successfully updated group for service template %s\n", stemplate.Name)
	} else if d.HasChange("gname") {
		group := d.Get("group").(string)
		gid, err := controller.Groups().ByName(group)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve group",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = stc.Chgrp(gid)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.Set("gid", gid)
		log.Printf("[INFO] Successfully updated group for service template %s\n", stemplate.Name)
	}

	if d.HasChange("uid") {
		user, err := controller.User(d.Get("uid").(int)).Info(true)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrive user informations",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = stc.Chown(d.Get("uid").(int), -1)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change user",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.Set("uname", user.Name)
		log.Printf("[INFO] Successfully updated owner for service template %s\n", stemplate.Name)
	} else if d.HasChange("uname") {
		uid, err := controller.Users().ByName(d.Get("uname").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve user",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = stc.Chown(uid, -1)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change user",
				Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.Set("uid", uid)
		log.Printf("[INFO] Successfully updated owner for service template %s\n", stemplate.Name)
	}

	return resourceOpennebulaServiceTemplateRead(ctx, d, meta)
}

// Helpers

func getServiceTemplateController(d *schema.ResourceData, meta interface{}) (*goca.STemplateController, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	var stc *goca.STemplateController

	if d.Id() != "" {
		id, err := strconv.ParseUint(d.Id(), 10, 0)
		if err != nil {
			return nil, err
		}
		stc = controller.STemplate(int(id))
	} else {
		return nil, fmt.Errorf("[ERROR] Service template ID cannot be found")
	}

	return stc, nil
}

func changeServiceTemplateGroup(d *schema.ResourceData, meta interface{}, stc *goca.STemplateController) error {
	config := meta.(*Configuration)
	controller := config.Controller
	var gid int
	var err error

	if d.Get("gname") != "" {
		gid, err = controller.Groups().ByName(d.Get("gname").(string))
		if err != nil {
			return err
		}
	} else {
		gid = d.Get("gid").(int)
	}

	err = stc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func changeServiceTemplateOwner(d *schema.ResourceData, meta interface{}, stc *goca.STemplateController) error {
	config := meta.(*Configuration)
	controller := config.Controller
	var uid int
	var err error

	if d.Get("uname") != "" {
		uid, err = controller.Users().ByName(d.Get("uname").(string))
		if err != nil {
			return err
		}
	} else {
		uid = d.Get("uid").(int)
	}

	err = stc.Chown(uid, -1)
	if err != nil {
		return err
	}

	return nil
}

func changeServiceTemplateName(d *schema.ResourceData, meta interface{}, stc *goca.STemplateController) error {
	if d.Get("name") != "" {
		err := stc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
	}

	return nil
}

func deepEqualIgnoreEmpty(old_template, new_template *srv_tmpl.ServiceTemplate) bool {
	old_map := structToMap(old_template)
	new_map := structToMap(new_template)

	for key, oldValue := range old_map {
		// Ignore empty fields since they are not returned by the OCA
		if isEmptyValue(reflect.ValueOf(oldValue)) {
			continue
		}

		// If the field is not present or not equal in the new template, return false
		newValue, ok := new_map[key]
		if !ok || !reflect.DeepEqual(oldValue, newValue) {
			return false
		}
	}

	return true
}

func structToMap(obj interface{}) map[string]interface{} {
	jsonBytes, _ := json.Marshal(obj)
	var mapObj map[string]interface{}
	json.Unmarshal(jsonBytes, &mapObj)
	return mapObj
}
