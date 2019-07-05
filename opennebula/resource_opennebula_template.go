package opennebula

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/kylelemons/godebug/pretty"
	"log"
	"strconv"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/template"
)

func resourceOpennebulaTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaTemplateCreate,
		Read:   resourceOpennebulaTemplateRead,
		Exists: resourceOpennebulaTemplateExists,
		Update: resourceOpennebulaTemplateUpdate,
		Delete: resourceOpennebulaTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the template",
			},
			"template": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of the template, in OpenNebula's XML or String format",
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
				Description: "ID of the user that will own the template",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the template",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the template",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the template",
			},
			"reg_time": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Registration time",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the Template, If empty, it uses caller group",
			},
		},
	}
}

func getTemplateController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.TemplateController, error) {
	controller := meta.(*goca.Controller)
	var tc *goca.TemplateController

	// Try to find the template by ID, if specified
	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Group Id (%s) is not an integer", d.Id())
		}
		tc = controller.Template(int(gid))
	}

	// Otherwise, try to find the template by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.Templates().ByName(d.Get("name").(string), args...)
		if err != nil {
			d.SetId("")
			return nil, fmt.Errorf("Could not find Template with name %s", d.Get("name").(string))
		}
		tc = controller.Template(gid)
	}

	return tc, nil
}

func changeTemplateGroup(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var gid int

	tc, err := getTemplateController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		gid, err = controller.Groups().ByName(d.Get("group").(string))
		if err != nil {
			return err
		}
	} else {
		gid = d.Get("gid").(int)
	}

	err = tc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	template, xmlerr := generateTemplateXML(d)
	if xmlerr != nil {
		return xmlerr
	}

	templateID, err := controller.Templates().Create(template)
	if err != nil {
		log.Printf("[ERROR] Template creation failed, error: %s", err)
		return err
	}

	tc := controller.Template(templateID)

	// add template information into Template
	err = tc.Update(d.Get("template").(string), 1)

	d.SetId(fmt.Sprintf("%v", templateID))

	// Change Permissions only if Permissions are set
	if perms, ok := d.GetOk("permissions"); ok {
		err = tc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			log.Printf("[ERROR] template permissions change failed, error: %s", err)
			return err
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err = changeTemplateGroup(d, meta)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaTemplateRead(d, meta)
}

func resourceOpennebulaTemplateRead(d *schema.ResourceData, meta interface{}) error {
	// Get requested template from all templates
	tc, err := getTemplateController(d, meta, -2, -1, -1)
	if err != nil {
		return err
	}

	template, err := tc.Info()
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", template.ID))
	d.Set("name", template.Name)
	d.Set("uid", template.UID)
	d.Set("gid", template.GID)
	d.Set("uname", template.UName)
	d.Set("gname", template.GName)
	d.Set("reg_time", template.RegTime)
	d.Set("permissions", permissionsUnixString(template.Permissions))

	// Get Human readable template information
	tpl := pretty.Sprint(template.Template)

	d.Set("template", tpl)

	return nil
}

func resourceOpennebulaTemplateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaTemplateRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	//Get Template
	tc, err := getTemplateController(d, meta)
	if err != nil {
		return err
	}
	template, err := tc.Info()
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		err := tc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
		// Update Name in internal struct
		template, err = tc.Info()
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated name for template %s\n", template.Name)
	}

	if d.HasChange("template") && d.Get("template") != "" {
		// replace the whole template instead of merging it with the existing one
		err = tc.Update(d.Get("template").(string), 0)
		if err != nil {
			return err
		}

		log.Printf("[INFO] Successfully updated template template %s\n", template.Name)

	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = tc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated Template %s\n", template.Name)
	}

	if d.HasChange("group") || d.HasChange("gid") {
		err = changeTemplateGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for Template %s\n", template.Name)
	}

	return nil
}

func resourceOpennebulaTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	tc, err := getTemplateController(d, meta)
	if err != nil {
		return err
	}

	err = tc.Delete()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted Template ID %s\n", d.Id())

	return nil
}

func generateTemplateXML(d *schema.ResourceData) (string, error) {
	name := d.Get("name").(string)

	template := &template.Template{
		Name: name,
	}

	w := &bytes.Buffer{}

	//Encode the Template schema to XML
	enc := xml.NewEncoder(w)
	//enc.Indent("", "  ")
	if err := enc.Encode(template); err != nil {
		return "", err
	}

	log.Printf("Template XML: %s", w.String())
	return w.String(), nil
}
