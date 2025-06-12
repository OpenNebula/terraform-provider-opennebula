package opennebula

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

const endpointFTemplate = "service_template"
const endpointFTemplateAction = "action"

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
				ForceNew:    false,
				Description: "Service Template body in json format",
				// Check JSON structure diffs, not binary diffs
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Custom deepEqual func to avoid empty fields #468
					return deepEqualIgnoreEmpty(old, new)
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
	var template map[string]interface{}

	if !config.isFlowConfigured() {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Flow client isn't configured",
			Detail:   fmt.Sprintf("Check flow_endpoint in the provider configuration"),
		})
		return diags
	}

	err := json.Unmarshal([]byte(d.Get("template").(string)), &template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse service template json description",
			Detail:   err.Error(),
		})
		return diags
	}

	body, err := getTemplateBody(template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve template body",
			Detail:   fmt.Sprintf("Error retrieving template body: %s", err),
		})
		return diags
	}

	response, err := applyTemplate(controller, endpointFTemplate, body)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the service template",
			Detail:   err.Error(),
		})
		return diags
	}

	respondeBody, err := getDocumentJSON(response)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse service template response",
			Detail:   fmt.Sprintf("Error parsing service template response: %s", err),
		})
		return diags
	}

	templateID, err := strconv.Atoi(respondeBody["ID"].(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template ID",
			Detail:   fmt.Sprintf("Response body: %v", respondeBody),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", templateID))
	stc := controller.STemplate(templateID)

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

	stc, err := getServiceTemplateController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the service template controller",
			Detail:   err.Error(),
		})
		return diags
	}

	templateDocument, err := templateRequest(controller, stc.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to read service template",
			Detail:   fmt.Sprintf("Error reading service template: %s", err),
		})
		return diags
	}
	templateJSON, err := getDocumentJSON(templateDocument)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse service template JSON",
			Detail:   fmt.Sprintf("Error parsing service template: %s", err),
		})
		return diags
	}

	templateID, err := strconv.Atoi(templateJSON["ID"].(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template ID",
			Detail:   fmt.Sprintf("Response body: %v", templateDocument),
		})
		return diags
	}

	name, err := getDocumentKey("NAME", templateJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template name",
			Detail:   fmt.Sprintf("Error retrieving service template name: %s", err),
		})
		return diags
	}

	uid, err := getDocumentKey("UID", templateJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template owner ID",
			Detail:   fmt.Sprintf("Error retrieving service template owner ID: %s", err),
		})
		return diags
	}
	if uid != nil {
		if uidStr, ok := uid.(string); ok {
			uidInt, err := strconv.Atoi(uidStr)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to convert service template owner ID",
					Detail:   fmt.Sprintf("Error converting service template owner ID: %s", err),
				})
				return diags
			}
			uid = uidInt
		}
	}

	gid, err := getDocumentKey("GID", templateJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template group ID",
			Detail:   fmt.Sprintf("Error retrieving service template group ID: %s", err),
		})
		return diags
	}
	if gid != nil {
		if gidStr, ok := gid.(string); ok {
			gidInt, err := strconv.Atoi(gidStr)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to convert service template group ID",
					Detail:   fmt.Sprintf("Error converting service template group ID: %s", err),
				})
				return diags
			}
			gid = gidInt
		}
	}

	uname, err := getDocumentKey("UNAME", templateJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template owner name",
			Detail:   fmt.Sprintf("Error retrieving service template owner name: %s", err),
		})
		return diags
	}

	gname, err := getDocumentKey("GNAME", templateJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template group name",
			Detail:   fmt.Sprintf("Error retrieving service template group name: %s", err),
		})
		return diags
	}

	permissions, err := getDocumentKey("PERMISSIONS", templateJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template permissions",
			Detail:   fmt.Sprintf("Error retrieving service template permissions: %s", err),
		})
		return diags
	}

	template, err := getDocumentKey("TEMPLATE", templateJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template body",
			Detail:   fmt.Sprintf("Error retrieving service template body: %s", err),
		})
		return diags
	}
	templateString, err := json.Marshal(template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to marshal service template",
			Detail:   fmt.Sprintf("Error marshaling service template: %s", err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", templateID))
	d.Set("name", name)
	d.Set("uid", uid)
	d.Set("gid", gid)
	d.Set("uname", uname)
	d.Set("gname", gname)
	d.Set("template", "{\"TEMPLATE\":"+string(templateString)+"}")
	err = d.Set("permissions", convertPermissions(permissions.(map[string]interface{})))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set attribute",
			Detail:   fmt.Sprintf("service template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

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

	_, err = templateRequest(controller, int(serviceTemplateID))
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaServiceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics
	var newTemplate map[string]interface{}

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

	if d.HasChange("template") {
		err := json.Unmarshal([]byte(d.Get("template").(string)), &newTemplate)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to parse service template json description",
				Detail:   err.Error(),
			})
			return diags
		}

		body, err := getTemplateBody(newTemplate)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve template body",
				Detail:   fmt.Sprintf("Error retrieving template body: %s", err),
			})
			return diags
		}
		// Apply the changes
		err = updateTemplateRequest(controller, stc.ID, body)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update service template",
				Detail:   fmt.Sprintf("Error updating service template: %s", err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated the template for service template ID %x\n", stc.ID)
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

		log.Printf("[INFO] Successfully updated name for service template ID %x\n", stc.ID)
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
		log.Printf("[INFO] Successfully updated Permissions for service template ID %x\n", stc.ID)
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
		log.Printf("[INFO] Successfully updated group for service template ID %d\n", stc.ID)
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
		log.Printf("[INFO] Successfully updated group for service template ID %d\n", stc.ID)
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
		log.Printf("[INFO] Successfully updated owner for service template ID %d\n", stc.ID)
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
		log.Printf("[INFO] Successfully updated owner for service template ID %d\n", stc.ID)
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

func deepEqualIgnoreEmpty(old_template, new_template string) bool {
	var old_map, new_map map[string]interface{}
	err := json.Unmarshal([]byte(old_template), &old_map)
	if err != nil {
		return false
	}
	err = json.Unmarshal([]byte(new_template), &new_map)
	if err != nil {
		return false
	}
	old_map_template, err := getTemplateBody(old_map)
	if err != nil {
		return false
	}
	new_map_template, err := getTemplateBody(new_map)
	if err != nil {
		return false
	}

	for key, oldValue := range old_map_template {
		if key == "registration_time" {
			continue
		}
		// If the field is not present or not equal in the new template, return false
		newValue, ok := new_map_template[key]
		if !ok || !reflect.DeepEqual(oldValue, newValue) {
			if key == "description" && oldValue == "" && newValue == nil {
				continue
			}
			return false
		}
	}

	// New parameters
	for key, newValue := range new_map_template {
		if key == "registration_time" {
			continue
		}
		oldValue, ok := old_map_template[key]
		if !ok || !reflect.DeepEqual(oldValue, newValue) {
			if key == "description" && newValue == "" && oldValue == nil {
				continue
			}
			return false
		}
	}

	return true
}

func templateRequest(c *goca.Controller, templateID int) (*goca.Response, error) {
	response, err := c.ClientFlow.HTTPMethod("GET", templateEndpoint(templateID))
	if err != nil {
		return nil, err
	}
	if !getReqStatusValue(response) {
		return nil, errors.New("failed to get template: " + response.Body())
	}
	return response, nil
}

func applyTemplate(c *goca.Controller, endpoint string, body map[string]interface{}) (*goca.Response, error) {
	response, err := c.ClientFlow.HTTPMethod("POST", endpoint, body)
	if err != nil {
		return nil, err
	}
	if !getReqStatusValue(response) {
		return nil, errors.New("failed to apply template: " + response.Body())
	}
	return response, nil
}

func getDocumentJSON(r *goca.Response) (map[string]interface{}, error) {
	docJSON, ok := r.BodyMap()["DOCUMENT"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("DOCUMENT key not found in response body")
	}

	return docJSON, nil
}

func getTemplateBody(template map[string]interface{}) (map[string]interface{}, error) {
	tmpl, ok := template["TEMPLATE"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid template structure: missing TEMPLATE")
	}

	body, ok := tmpl["BODY"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid template structure: missing BODY")
	}
	return body, nil
}

func templateEndpoint(templateID int) string {
	return fmt.Sprintf("%s/%s", endpointFTemplate, strconv.Itoa(templateID))
}

func templateEndpointAction(templateID int) string {
	return fmt.Sprintf("%s/%s", templateEndpoint(templateID), endpointFTemplateAction)
}

func getDocumentKey(key string, doc map[string]interface{}) (interface{}, error) {
	if doc == nil {
		return "", fmt.Errorf("document is nil")
	}
	if value, exists := doc[key]; exists {
		return value, nil
	}
	return "", fmt.Errorf("key %s not found in document", key)
}

func convertPermissions(permissions map[string]interface{}) string {
	parse := func(key string) int8 {
		s, ok := permissions[key].(string)
		if !ok {
			return 0
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		if n < math.MinInt8 || n > math.MaxInt8 {
			return 0
		}
		return int8(n)
	}

	return permissionsUnixString(shared.Permissions{
		OwnerU: parse("OWNER_U"),
		OwnerM: parse("OWNER_M"),
		OwnerA: parse("OWNER_A"),
		GroupU: parse("GROUP_U"),
		GroupM: parse("GROUP_M"),
		GroupA: parse("GROUP_A"),
		OtherU: parse("OTHER_U"),
		OtherM: parse("OTHER_M"),
		OtherA: parse("OTHER_A"),
	})
}

func updateTemplateRequest(c *goca.Controller, templateID int,
	newTemplate map[string]interface{}) error {
	newTemplateString, err := json.Marshal(newTemplate)
	if err != nil {
		return fmt.Errorf("failed to marshal new template: %s", err)
	}
	action := make(map[string]interface{})
	action["action"] = map[string]interface{}{
		"perform": "update",
		"params": map[string]interface{}{
			"append":        false,
			"template_json": string(newTemplateString),
		},
	}
	_, err = applyTemplate(c, templateEndpointAction(templateID), action)
	return err
}

func getReqStatusValue(response *goca.Response) bool {
	if response == nil {
		return false
	}
	defer func() {
		if recover() != nil {
			log.Printf("[ERROR] Failed to retrieve status from response: %v", response)
		}
	}()
	v := reflect.ValueOf(response).Elem()
	statusField := v.FieldByName("status")

	if statusField.IsValid() && statusField.Kind() == reflect.Bool {
		return statusField.Bool()
	}
	return false
}
