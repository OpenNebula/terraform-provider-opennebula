package opennebula

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

func resourceOpennebulaService() *schema.Resource {
	return &schema.Resource{
		Create:        resourceOpennebulaServiceCreate,
		Read:          resourceOpennebulaServiceRead,
		//Exists:        resourceOpennebulaVirtualMachineExists,
		//Update:        resourceOpennebulaVirtualMachineUpdate,
		Delete:        resourceOpennebulaServiceDelete,
		//CustomizeDiff: resourceVMCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the Service. If empty, defaults to 'templatename-<vmid>'",
			},
			"template_id": {
				Type:        schema.TypeInt,
				Optional:    true,
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
				Description: "ID of the user that will own the Service",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the Service",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the Service",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the Service",
			},
			"state": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Current state of the Service",
			},
		},
	}
}

func resourceOpennebulaServiceCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	var err error
	var serviceID int

	if v, ok := d.GetOkExists("template_id"); ok {
		// if template id is set, instantiate a Service from this template
		tc := controller.STemplate(v.(int))

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
	} else {
		return fmt.Errorf("A valid template_id is mandatory.")
	}

	d.SetId(fmt.Sprintf("%v", serviceID))
	sc := controller.Service(serviceID)

	//Set the permissions on the VM if it was defined, otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = sc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			log.Printf("[ERROR] template permissions change failed, error: %s", err)
			return err
		}
	}

	if d.Get("gname") != "" || d.Get("gid") != "" {
		err = changeServiceGroup(d, meta, sc)
		if err != nil {
			return err
		}
	}

	if d.Get("uname") != "" || d.Get("uid") != "" {
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

	return nil
}

func resourceOpennebulaServiceDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// Helpers

func getServiceController(d *schema.ResourceData, meta interface{}) (*goca.ServiceController, error) {
	controller := meta.(*goca.Controller)
	var sc *goca.ServiceController

	// Try to find the VM by ID, if specified
	if d.Id() != "" {
		id, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		sc = controller.Service(int(id))
	} else {
		return nil, fmt.Errorf("[ERROR] Template ID cannot be found")
	}

	return sc, nil
}

func changeServiceGroup(d *schema.ResourceData, meta interface{}, sc *goca.ServiceController) error {
	controller := meta.(*goca.Controller)
	var gid int
	var err error

	if d.Get("gname") != "" {
		gid, err = controller.Groups().ByName(d.Get("group").(string))
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
	controller := meta.(*goca.Controller)
	var uid int
	var err error

	if d.Get("uname") != "" {
		uid, err = controller.Users().ByName(d.Get("group").(string))
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
