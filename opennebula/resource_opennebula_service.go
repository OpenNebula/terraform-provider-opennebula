package opennebula

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/service"
)

func resourceOpennebulaService() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaServiceCreate,
		Read:   resourceOpennebulaServiceRead,
		Exists: resourceOpennebulaServiceExists,
		Update: resourceOpennebulaServiceUpdate,
		Delete: resourceOpennebulaServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				ForceNew:    true,
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
				Description: "Map with the role dinamically generated information",
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

func resourceOpennebulaServiceCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller

	var err error
	var serviceID int

	// if template id is set, instantiate a Service from this template
	tc := controller.STemplate(d.Get("template_id").(int))

	var extra_template = ""
	if v, ok := d.GetOk("extra_template"); ok {
		extra_template = v.(string)
	}

	// Instantiate template
	service, err := tc.Instantiate(extra_template)
	if err != nil {
		return err
	}

	serviceID = service.ID

	d.SetId(fmt.Sprintf("%v", serviceID))
	sc := controller.Service(serviceID)

	//Set the permissions on the Service if it was defined, otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = sc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			log.Printf("[ERROR] service permissions change failed, error: %s", err)
			return err
		}
	}

	if _, ok := d.GetOkExists("gid"); d.Get("gname") != "" || ok {
		err = changeServiceGroup(d, meta, sc)
		if err != nil {
			return err
		}
	}

	if _, ok := d.GetOkExists("uid"); d.Get("uname") != "" || ok {
		err = changeServiceOwner(d, meta, sc)
		if err != nil {
			return err
		}
	}

	if d.Get("name") != "" {
		err = changeServiceName(d, meta, sc)
		if err != nil {
			return err
		}
	}

	_, err = waitForServiceState(d, meta, "running")
	if err != nil {
		service, _ := sc.Info()

		svState := service.Template.Body.StateRaw
		return fmt.Errorf(
			"Error waiting for Service (%s) to be in state RUNNING: %s (state: %v)", d.Id(), err, svState)
	}

	return resourceOpennebulaServiceRead(d, meta)
}

func resourceOpennebulaServiceRead(d *schema.ResourceData, meta interface{}) error {
	sc, err := getServiceController(d, meta)
	if err != nil {
		return err
	}

	sv, err := sc.Info()
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", sv.ID))
	d.Set("name", sv.Name)
	d.Set("uid", sv.UID)
	d.Set("gid", sv.GID)
	d.Set("uname", sv.UName)
	d.Set("gname", sv.GName)
	d.Set("state", sv.Template.Body.StateRaw)
	err = d.Set("permissions", permissionsUnixString(*sv.Permissions))
	if err != nil {
		return err
	}

	// Retrieve networks
	var networks = make(map[string]int)
	for _, val := range sv.Template.Body.NetworksVals {
		for k, v := range val {
			networks[k] = int(v.(map[string]interface{})["id"].(float64))
		}
	}
	d.Set("networks", networks)

	// Retrieve roles
	var roles []map[string]interface{}
	for _, role := range sv.Template.Body.Roles {
		role_tf := make(map[string]interface{})
		role_tf["name"] = role.Name
		role_tf["cardinality"] = role.Cardinality
		role_tf["state"] = role.StateRaw

		var nodes_ids []int
		for _, node := range role.Nodes {
			nodes_ids = append(nodes_ids, node.VMInfo.VM.ID)
		}

		role_tf["nodes"] = nodes_ids

		roles = append(roles, role_tf)
	}
	d.Set("roles", roles)

	return nil
}

func resourceOpennebulaServiceDelete(d *schema.ResourceData, meta interface{}) error {
	err := resourceOpennebulaServiceRead(d, meta)
	if err != nil || d.Id() == "" {
		return err
	}

	//Get Service
	sc, err := getServiceController(d, meta)
	if err != nil {
		return err
	}

	if err = sc.Delete(); err != nil {
		if err = sc.Recover(true); err != nil {
			return err
		}
	}

	_, err = waitForServiceState(d, meta, "done")
	if err != nil {
		service, _ := sc.Info()

		svState := service.Template.Body.StateRaw
		return fmt.Errorf(
			"Error waiting for Service (%s) to be in state DONE: %s (state: %v)", d.Id(), err, svState)
	}

	log.Printf("[INFO] Successfully terminated service\n")
	return nil
}

func resourceOpennebulaServiceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaServiceRead(d, meta)
	// a terminated Service is in state 5 (DONE)
	if err != nil || d.Id() == "" || d.Get("state").(int) == 5 {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller

	//Get Service controller
	sc, err := getServiceController(d, meta)
	if err != nil {
		return err
	}

	service, err := sc.Info()
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		err := sc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}

		service, err := sc.Info()
		log.Printf("[INFO] Successfully updated name (%s) for Service ID %x\n", service.Name, service.ID)
	}

	if d.HasChange("permissions") && d.Get("permissions") != "" {
		if perms, ok := d.GetOk("permissions"); ok {
			err = sc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated Permissions for Service %s\n", service.Name)
	}

	if d.HasChange("gid") {
		gid := d.Get("gid").(int)
		group, err := controller.Group(gid).Info(true)
		if err != nil {
			return fmt.Errorf("Can't find a group with ID `%d`: %s", gid, err)
		}
		err = sc.Chgrp(gid)
		if err != nil {
			return err
		}

		d.Set("gname", group.Name)
		log.Printf("[INFO] Successfully updated group for Service %s\n", service.Name)
	} else if d.HasChange("gname") {
		group := d.Get("group").(string)
		gid, err := controller.Groups().ByName(group)
		if err != nil {
			return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
		}
		err = sc.Chgrp(gid)
		if err != nil {
			return err
		}

		d.Set("gid", gid)
		log.Printf("[INFO] Successfully updated group for Service %s\n", service.Name)
	}

	if d.HasChange("uid") {
		user, err := controller.User(d.Get("uid").(int)).Info(true)
		if err != nil {
			return err
		}
		err = sc.Chown(d.Get("uid").(int), -1)
		if err != nil {
			return err
		}

		d.Set("uname", user.Name)
		log.Printf("[INFO] Successfully updated owner for Service %s\n", service.Name)
	} else if d.HasChange("uname") {
		uid, err := controller.Users().ByName(d.Get("uname").(string))
		if err != nil {
			return err
		}
		err = sc.Chown(uid, -1)
		if err != nil {
			return err
		}

		d.Set("uid", uid)
		log.Printf("[INFO] Successfully updated owner for Service %s\n", service.Name)
	}

	return resourceOpennebulaServiceRead(d, meta)
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

func waitForServiceState(d *schema.ResourceData, meta interface{}, state string) (interface{}, error) {
	var service *service.Service
	var err error

	//Get Service controller
	sc, err := getServiceController(d, meta)
	if err != nil {
		return service, err
	}

	log.Printf("Waiting for Service (%s) to be in state %s", d.Id(), state)

	stateConf := &resource.StateChangeConf{
		Pending: []string{"anythingelse"}, Target: []string{state},
		Refresh: func() (interface{}, string, error) {
			log.Println("Refreshing Service state...")
			if d.Id() != "" {
				//Get Service controller
				sc, err = getServiceController(d, meta)
				if err != nil {
					return service, "", fmt.Errorf("Could not find Service by ID %s", d.Id())
				}
			}

			service, err = sc.Info()
			if err != nil {
				if strings.Contains(err.Error(), "Error getting") {
					if state == "done" {
						return service, "done", nil // DONE == notfound for ONE > 5.12
					} else {
						return service, "notfound", nil
					}
				}
				return service, "", err
			}
			svState := service.Template.Body.StateRaw
			if err != nil {
				if strings.Contains(err.Error(), "Error getting") {
					return service, "notfound", nil
				}
				return service, "", err
			}
			log.Printf("Service %v is currently in state %v", service.ID, svState)
			if svState == 2 {
				return service, "running", nil
			} else if svState == 4 {
				return service, "warning", fmt.Errorf("Service ID %s entered warning state", d.Id())
			} else if svState == 5 {
				return service, "done", nil
			} else if svState == 6 {
				return service, "failed_undeploying", fmt.Errorf("Service ID %s entered failed_undeploying state", d.Id())
			} else if svState == 7 {
				return service, "failed_deploying", fmt.Errorf("Service ID %s entered failed_deploying state", d.Id())
			} else if svState == 9 {
				return service, "failed_scaling", fmt.Errorf("Service ID %s entered failed_scaling state", d.Id())
			} else if svState == 10 {
				return service, "cooldown", nil
			} else {
				return service, "anythingelse", nil
			}
		},
		Timeout:    3 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}
