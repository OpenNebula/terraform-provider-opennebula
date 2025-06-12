package opennebula

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	ver "github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

const endpointFService = "service"

var (
	defaultServiceTimeoutMin = 20
	defaultServiceTimeout    = time.Duration(defaultVMTimeoutMin) * time.Minute
)

func resourceOpennebulaService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaServiceCreate,
		ReadContext:   resourceOpennebulaServiceRead,
		Exists:        resourceOpennebulaServiceExists,
		UpdateContext: resourceOpennebulaServiceUpdate,
		DeleteContext: resourceOpennebulaServiceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultServiceTimeout),
			Delete: schema.DefaultTimeout(defaultServiceTimeout),
			Update: schema.DefaultTimeout(defaultServiceTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the Service",
			},
			"template_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the Service template to use",
			},
			"extra_template": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Extra template information in json format to be added to the service template during instantiate.",
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the service (in Unix format, owner-group-other, use-manage-admin)",
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
				Description: "ID of the user that will own the Service",
			},
			"gid": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "ID of the group that will own the Service",
			},
			"uname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the user that will own the Service",
			},
			"gname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the group that will own the Service",
			},
			"state": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Current state of the Service",
			},
			"networks": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Map with the service networks names as key and id as value",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"roles": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Map with the role dynamically generated information",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cardinality": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Cardinality of the role",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the Role",
						},
						"nodes": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of role nodes",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"state": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Current state of the role",
						},
					},
				},
			},
		},
	}
}

func resourceOpennebulaServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	var err error
	var serviceID int

	// if template id is set, instantiate a Service from this template
	templateID, ok := d.Get("template_id").(int)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get template ID",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	var extra_template = ""
	if v, ok := d.GetOk("extra_template"); ok {
		extra_template = v.(string)
	}

	// Instantiate template
	response, err := instantiateTemplate(templateID, extra_template, controller)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to instantiate service",
			Detail:   err.Error(),
		})
		return diags
	}
	responseBody, err := getDocumentJSON(response)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse response",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	log.Printf("[INFO] Successfully instantiated service template %d with response: %s\n", templateID, response.BodyMap())
	serviceID, err = strconv.Atoi(responseBody["ID"].(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service template ID",
			Detail:   fmt.Sprintf("Response body: %v", responseBody),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", serviceID))
	sc := controller.Service(serviceID)

	//Set the permissions on the Service if it was defined, otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = sc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if _, ok := d.GetOkExists("gid"); d.Get("gname") != "" || ok {
		err = changeServiceGroup(d, meta, sc)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if _, ok := d.GetOkExists("uid"); d.Get("uname") != "" || ok {
		err = changeServiceOwner(d, meta, sc)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change owner",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("name") != "" {
		err = changeServiceName(d, meta, sc)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	_, err = waitForServiceState(ctx, d, meta, "running", timeout, controller)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to wait service to be in RUNNING state",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return resourceOpennebulaServiceRead(ctx, d, meta)
}

func resourceOpennebulaServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

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

	sc, err := getServiceController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the service controller",
			Detail:   err.Error(),
		})
		return diags
	}

	response, err := getServiceInfo(controller, sc.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve service information",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	serviceJSON, err := getDocumentJSON(response)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse service document",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	serviceBody, err := getTemplateBody(serviceJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service template body",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	id, err := getDocumentKey("ID", serviceJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service ID",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	uid, err := getDocumentKey("UID", serviceJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service UID",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	if uid != nil {
		if uidStr, ok := uid.(string); ok {
			uidInt, err := strconv.Atoi(uidStr)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to convert service owner ID",
					Detail:   fmt.Sprintf("Error converting service owner ID: %s", err),
				})
				return diags
			}
			uid = uidInt
		}
	}

	gid, err := getDocumentKey("GID", serviceJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service GID",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	if gid != nil {
		if gidStr, ok := gid.(string); ok {
			gidInt, err := strconv.Atoi(gidStr)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to convert service group ID",
					Detail:   fmt.Sprintf("Error converting service group ID: %s", err),
				})
				return diags
			}
			gid = gidInt
		}
	}

	uname, err := getDocumentKey("UNAME", serviceJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service UNAME",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	gname, err := getDocumentKey("GNAME", serviceJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service GNAME",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	name, err := getDocumentKey("NAME", serviceJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service name",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	permissions, err := getDocumentKey("PERMISSIONS", serviceJSON)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service permissions",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	state, err := getDocumentKey("state", serviceBody)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get service state",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", id))
	d.Set("name", name)
	d.Set("uid", uid)
	d.Set("gid", gid)
	d.Set("uname", uname)
	d.Set("gname", gname)
	d.Set("state", state)
	err = d.Set("permissions", convertPermissions(permissions.(map[string]interface{})))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set attribute",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	networksVal, _ := getDocumentKey("networks_values", serviceBody)
	// Retrieve networks
	var networks = make(map[string]int)
	if networksVal != nil {
		networksArray, ok := networksVal.([]interface{})
		if ok && len(networksArray) > 0 {
			for _, val := range networksVal.([]interface{}) {
				for k, v := range val.(map[string]interface{}) {
					if v == nil {
						diags = append(diags, diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Network ID is nil",
							Detail:   fmt.Sprintf("service (ID: %s): Network ID is nil", d.Id()),
						})
						return diags
					}
					idInterface, ok := v.(map[string]interface{})["id"]
					if !ok {
						diags = append(diags, diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Network ID is missing",
							Detail:   fmt.Sprintf("service (ID: %s): Network ID is missing", d.Id()),
						})
						return diags
					}
					networkID, err := ParseIntFromInterface(idInterface)
					if err != nil {
						diags = append(diags, diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Failed to parse network ID",
							Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
						})
						return diags
					}
					networks[k] = networkID
				}
			}
		}
	}
	d.Set("networks", networks)

	// Retrieve roles
	rolesBody, _ := getDocumentKey("roles", serviceBody)
	var roles []map[string]interface{}
	for _, role := range rolesBody.([]interface{}) {
		role_tf := make(map[string]interface{})
		name, err := getDocumentKey("name", role.(map[string]interface{}))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to get role name",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		cardinality, err := getDocumentKey("cardinality", role.(map[string]interface{}))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to get role cardinality",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		stateRaw, err := getDocumentKey("state", role.(map[string]interface{}))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to get role state",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		role_tf["name"] = name
		role_tf["cardinality"] = cardinality
		role_tf["state"] = stateRaw

		var nodes_ids []int
		nodes, err := getDocumentKey("nodes", role.(map[string]interface{}))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to get role state",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		for _, node := range nodes.([]interface{}) {
			vmInfo, err := getDocumentKey("vm_info", node.(map[string]interface{}))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to get VM info",
					Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
			vmInfoVM, err := getDocumentKey("VM", vmInfo.(map[string]interface{}))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to get VM info",
					Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
			vmID, err := getDocumentKey("ID", vmInfoVM.(map[string]interface{}))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to get VM ID",
					Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
			if vmIDStr, ok := vmID.(string); ok {
				if id, err := strconv.Atoi(vmIDStr); err == nil {
					nodes_ids = append(nodes_ids, id)
				} else {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to convert VM ID to int",
						Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
		}

		role_tf["nodes"] = nodes_ids
		roles = append(roles, role_tf)
	}

	d.Set("roles", roles)

	return nil
}

func resourceOpennebulaServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

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

	//Get Service
	sc, err := getServiceController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the service controller",
			Detail:   err.Error(),
		})
		return diags
	}

	if err = sc.Delete(); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		if err = sc.Recover(true); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to recover",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	_, err = waitForServiceState(ctx, d, meta, "done", timeout, controller)
	if err != nil && !strings.Contains(err.Error(), "failed to get service") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to wait service to be in DONE state",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully terminated service\n")
	return nil
}

func resourceOpennebulaServiceExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	serviceID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = getServiceInfo(controller, int(serviceID))
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

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
	sc, err := getServiceController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the service controller",
			Detail:   err.Error(),
		})
		return diags
	}

	response, err := getServiceInfo(controller, sc.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if d.HasChange("name") {
		err := sc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change name",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name for Service ID %d\n", sc.ID)
	}

	if d.HasChange("permissions") && d.Get("permissions") != "" {
		if perms, ok := d.GetOk("permissions"); ok {
			err = sc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change permissions",
					Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated Permissions for Service %d\n", sc.ID)
	}

	if d.HasChange("gid") {
		gid := d.Get("gid").(int)
		group, err := controller.Group(gid).Info(true)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve informations",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = sc.Chgrp(gid)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.Set("gname", group.Name)
		log.Printf("[INFO] Successfully updated group for Service %d\n", sc.ID)
	} else if d.HasChange("gname") {
		group := d.Get("group").(string)
		gid, err := controller.Groups().ByName(group)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve group",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = sc.Chgrp(gid)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.Set("gid", gid)
		log.Printf("[INFO] Successfully updated group for Service %d\n", sc.ID)
	}

	if d.HasChange("uid") {
		user, err := controller.User(d.Get("uid").(int)).Info(true)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed retrieve user",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = sc.Chown(d.Get("uid").(int), -1)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change user",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.Set("uname", user.Name)
		log.Printf("[INFO] Successfully updated owner for Service %d\n", sc.ID)
	} else if d.HasChange("uname") {
		uid, err := controller.Users().ByName(d.Get("uname").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve user",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = sc.Chown(uid, -1)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change user",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.Set("uid", uid)
		log.Printf("[INFO] Successfully updated owner for Service %d\n", sc.ID)
	}

	if d.HasChange("extra_template") {
		extra_template := make(map[string]interface{})
		if v, ok := d.GetOk("extra_template"); ok {
			if err := json.Unmarshal([]byte(v.(string)), &extra_template); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to parse extra template",
					Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		type roleDesc struct {
			name           string
			oldCardinality int
			newCardinality int
			wantScale      bool
		}

		desc := []roleDesc{}
		serviceJSON, err := getDocumentJSON(response)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to parse service document",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		serviceBody, err := getTemplateBody(serviceJSON)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to get service template body",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		roles, err := getDocumentKey("roles", serviceBody)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to get roles from service template",
				Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		for _, role := range roles.([]interface{}) {
			roleName, err := getDocumentKey("name", role.(map[string]interface{}))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to get role name",
					Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

			roleCardinality, err := getDocumentKey("cardinality", role.(map[string]interface{}))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to get role cardinality",
					Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
			desc = append(desc, roleDesc{
				name:           roleName.(string),
				oldCardinality: int(roleCardinality.(float64)),
			})
		}

		if roles, ok := extra_template["roles"]; ok {
			for k := 0; k < len(desc) && k < len(roles.([]interface{})); k++ {
				if v, ok := roles.([]interface{})[k].(map[string]interface{})["cardinality"]; ok {
					desc[k].newCardinality = int(v.(float64))
					desc[k].wantScale = desc[k].newCardinality != desc[k].oldCardinality
				}
			}
		}

		minVersion, _ := ver.NewVersion("6.8.0")

		for _, v := range desc {
			if v.wantScale {
				if config.OneVersion.LessThan(minVersion) {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Role scaling is unsupported for this environment",
						Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
					})
					return diags
				}

				if err := sc.Scale(v.name, v.newCardinality, false); err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to scale role",
						Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
					})
					return diags
				}

				timeout := d.Timeout(schema.TimeoutUpdate)
				if _, err := waitForServiceState(ctx, d, meta, "running", timeout, controller); err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to wait service to be in RUNNING state",
						Detail:   fmt.Sprintf("service (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
		}

		log.Printf("[INFO] Successfully scaled roles of Service %d\n", sc.ID)
	}

	return resourceOpennebulaServiceRead(ctx, d, meta)
}

// Helpers

func getServiceController(d *schema.ResourceData, meta interface{}) (*goca.ServiceController, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	var sc *goca.ServiceController

	if d.Id() != "" {
		id, err := strconv.ParseUint(d.Id(), 10, 0)
		if err != nil {
			return nil, err
		}
		sc = controller.Service(int(id))
	} else {
		return nil, fmt.Errorf("[ERROR] Service ID cannot be found")
	}

	return sc, nil
}

func changeServiceGroup(d *schema.ResourceData, meta interface{}, sc *goca.ServiceController) error {
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

	err = sc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func changeServiceOwner(d *schema.ResourceData, meta interface{}, sc *goca.ServiceController) error {
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

	err = sc.Chown(uid, -1)
	if err != nil {
		return err
	}

	return nil
}

func changeServiceName(d *schema.ResourceData, meta interface{}, sc *goca.ServiceController) error {
	if d.Get("name") != "" {
		err := sc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForServiceState(ctx context.Context, d *schema.ResourceData, meta interface{}, state string, timeout time.Duration, c *goca.Controller) (interface{}, error) {
	var svState = -1
	var err error

	//Get Service controller
	sc, err := getServiceController(d, meta)
	if err != nil {
		return svState, err
	}

	log.Printf("Waiting for Service (%s) to be in state %s", d.Id(), state)

	stateConf := &resource.StateChangeConf{
		Pending: []string{"anythingelse", "cooldown"},
		Target:  []string{state},
		Refresh: func() (interface{}, string, error) {
			log.Println("Refreshing Service state...")
			if d.Id() != "" {
				//Get Service controller
				sc, err = getServiceController(d, meta)
				if err != nil {
					return svState, "", fmt.Errorf("Could not find Service by ID %s", d.Id())
				}
			}

			response, err := getServiceInfo(c, sc.ID)
			if err != nil {
				if strings.Contains(err.Error(), "Error getting") {
					if state == "done" {
						return svState, "done", nil // DONE == notfound for ONE > 5.12
					} else {
						return svState, "notfound", nil
					}
				}
				return svState, "", err
			}
			serviceDocument, err := getDocumentJSON(response)
			if err != nil {
				return svState, "", fmt.Errorf("Could not get template document for Service ID %s: %s", d.Id(), err)
			}
			serviceBody, err := getTemplateBody(serviceDocument)
			if err != nil {
				return svState, "", fmt.Errorf("Could not get template body for Service ID %s: %s", d.Id(), err)
			}

			state, err := getDocumentKey("state", serviceBody)
			if err != nil {
				return svState, "", fmt.Errorf("Could not get state from Service template body for Service ID %s: %s", d.Id(), err)
			}

			svState = int(state.(float64))
			log.Printf("Service %v is currently in state %v", d.Id(), svState)
			if svState == 2 {
				return svState, "running", nil
			} else if svState == 4 {
				return svState, "warning", fmt.Errorf("Service ID %s entered warning state", d.Id())
			} else if svState == 5 {
				return svState, "done", nil
			} else if svState == 6 {
				return svState, "failed_undeploying", fmt.Errorf("Service ID %s entered failed_undeploying state", d.Id())
			} else if svState == 7 {
				return svState, "failed_deploying", fmt.Errorf("Service ID %s entered failed_deploying state", d.Id())
			} else if svState == 9 {
				return svState, "failed_scaling", fmt.Errorf("Service ID %s entered failed_scaling state", d.Id())
			} else if svState == 10 {
				return svState, "cooldown", nil
			} else {
				return svState, "anythingelse", nil
			}
		},
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForStateContext(ctx)
}

func instantiateTemplate(templateID int, extraTmpl string, c *goca.Controller) (*goca.Response, error) {
	url := templateEndpointAction(templateID)
	action := make(map[string]interface{})
	args := make(map[string]interface{})
	params := make(map[string]interface{})

	if extraTmpl != "" {
		err := json.Unmarshal([]byte(extraTmpl), &args)
		if err != nil {
			return nil, err
		}

		params["merge_template"] = args
	}

	action["action"] = map[string]interface{}{
		"perform": "instantiate",
		"params":  params,
	}

	response, err := c.ClientFlow.HTTPMethod("POST", url, action)
	if err != nil {
		return nil, err
	}
	if !getReqStatusValue(response) {
		return nil, errors.New("failed to update service: " + strconv.Itoa(templateID))
	}
	return response, nil
}

func getServiceInfo(c *goca.Controller, serviceID int) (*goca.Response, error) {
	response, err := c.ClientFlow.HTTPMethod("GET", serviceEndpoint(serviceID))
	if err != nil {
		return nil, err
	}
	if !getReqStatusValue(response) {
		return nil, errors.New("failed to get service: " + strconv.Itoa(serviceID))
	}
	return response, nil
}

func serviceEndpoint(serviceID int) string {
	return fmt.Sprintf("%s/%s", endpointFService, strconv.Itoa(serviceID))
}
