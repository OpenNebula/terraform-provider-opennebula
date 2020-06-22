package opennebula

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	srv_tmpl "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/service_template"
)

func resourceOpennebulaServiceTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaServiceTemplateCreate,
		Read:   resourceOpennebulaServiceTemplateRead,
		Exists: resourceOpennebulaServiceTemplateExists,
		Update: resourceOpennebulaServiceTemplateUpdate,
		Delete: resourceOpennebulaServiceTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceOpennebulaServiceTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	var err error

	if _, ok := d.GetOk("name"); !ok {
		return fmt.Errorf("The Name is mandatory for creating a service template.")
	}

	if _, ok := d.GetOk("template"); !ok {
		return fmt.Errorf("The service template body is mandatory.")
	}

	// Marshall the json
	stemplate := &srv_tmpl.ServiceTemplate{}
	err = json.Unmarshal([]byte(d.Get("template").(string)), stemplate)
	if err != nil {
		return err
	}

	err = controller.STemplates().Create(stemplate)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", stemplate.ID))
	stc := controller.STemplate(stemplate.ID)

	// Set the permissions on the service template if it was defined,
	// otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = stc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			log.Printf("[ERROR] template permissions change failed, error: %s", err)
			return err
		}
	}

	if _, ok := d.GetOkExists("gid"); d.Get("gname") != "" || ok {
		err = changeServiceTemplateGroup(d, meta, stc)
		if err != nil {
			return err
		}
	}

	if _, ok := d.GetOkExists("uid"); d.Get("uname") != "" || ok {
		err = changeServiceTemplateOwner(d, meta, stc)
		if err != nil {
			return err
		}
	}

	if d.Get("name") != "" {
		err = changeServiceTemplateName(d, meta, stc)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaServiceTemplateRead(d, meta)
}

func resourceOpennebulaServiceTemplateRead(d *schema.ResourceData, meta interface{}) error {
	stc, err := getServiceTemplateController(d, meta)
	if err != nil {
		return err
	}

	st, err := stc.Info()
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", st.ID))
	d.Set("name", st.Name)
	d.Set("uid", st.UID)
	d.Set("gid", st.GID)
	d.Set("uname", st.UName)
	d.Set("gname", st.GName)
	err = d.Set("permissions", permissionsUnixString(*st.Permissions))
	if err != nil {
		return err
	}

	// Get service.Template as map
	tmpl_byte, err := json.Marshal(st.Template)
	if err != nil {
		return err
	}
	d.Set("template", "{\"TEMPLATE\":"+string(tmpl_byte)+"}")

	return nil
}

func resourceOpennebulaServiceTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	err := resourceOpennebulaServiceTemplateRead(d, meta)
	if err != nil || d.Id() == "" {
		return err
	}

	//Get Service
	stc, err := getServiceTemplateController(d, meta)
	if err != nil {
		return err
	}

	if err = stc.Delete(); err != nil {
		return err
	}

	log.Printf("[INFO] Successfully terminated service template\n")

	return nil
}

func resourceOpennebulaServiceTemplateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaServiceTemplateRead(d, meta)
	if err != nil {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaServiceTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	// Enable partial state mode
	d.Partial(true)

	//Get Service controller
	stc, err := getServiceTemplateController(d, meta)
	if err != nil {
		return err
	}

	stemplate, err := stc.Info()
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		err := stc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}

		stemplate, err := stc.Info()
		d.SetPartial("name")
		log.Printf("[INFO] Successfully updated name (%s) for service template ID %x\n", stemplate.Name, stemplate.ID)
	}

	if d.HasChange("permissions") && d.Get("permissions") != "" {
		if perms, ok := d.GetOk("permissions"); ok {
			err = stc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		d.SetPartial("permissions")
		log.Printf("[INFO] Successfully updated Permissions for service template %s\n", stemplate.Name)
	}

	if d.HasChange("gid") {
		group, err := controller.Group(d.Get("gid").(int)).Info(true)
		if err != nil {
			return err
		}
		err = stc.Chgrp(d.Get("gid").(int))
		if err != nil {
			return err
		}

		d.Set("gname", group.Name)
		d.SetPartial("gname")
		d.SetPartial("gid")
		log.Printf("[INFO] Successfully updated group for service template %s\n", stemplate.Name)
	} else if d.HasChange("gname") {
		gid, err := controller.Groups().ByName(d.Get("gname").(string))
		if err != nil {
			return err
		}
		err = stc.Chgrp(gid)
		if err != nil {
			return err
		}

		d.Set("gid", gid)
		d.SetPartial("gid")
		d.SetPartial("gname")
		log.Printf("[INFO] Successfully updated group for service template %s\n", stemplate.Name)
	}

	if d.HasChange("uid") {
		user, err := controller.User(d.Get("uid").(int)).Info(true)
		if err != nil {
			return err
		}
		err = stc.Chown(d.Get("uid").(int), -1)
		if err != nil {
			return err
		}

		d.Set("uname", user.Name)
		d.SetPartial("uname")
		d.SetPartial("uid")
		log.Printf("[INFO] Successfully updated owner for service template %s\n", stemplate.Name)
	} else if d.HasChange("uname") {
		uid, err := controller.Users().ByName(d.Get("uname").(string))
		if err != nil {
			return err
		}
		err = stc.Chown(uid, -1)
		if err != nil {
			return err
		}

		d.Set("uid", uid)
		d.SetPartial("uid")
		d.SetPartial("uname")
		log.Printf("[INFO] Successfully updated owner for service template %s\n", stemplate.Name)
	}

	// We succeeded, disable partial mode. This causes Terraform to save
	// save all fields again.
	d.Partial(false)

	return nil
}

// Helpers

func getServiceTemplateController(d *schema.ResourceData, meta interface{}) (*goca.STemplateController, error) {
	controller := meta.(*goca.Controller)
	var stc *goca.STemplateController

	if d.Id() != "" {
		id, err := strconv.ParseUint(d.Id(), 10, 64)
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
	controller := meta.(*goca.Controller)
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
	controller := meta.(*goca.Controller)
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
