package opennebula

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
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
				Optional:    true,
				Computed:    true,
				Description: "Description of the template, in OpenNebula's XML or String format",
				Deprecated:  "use other schema sections instead.",
			},
			"cpu":     cpuSchema(),
			"vcpu":    vcpuSchema(),
			"memory":  memorySchema(),
			"context": contextSchema(),
			"cpumodel": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Definition of CPU Model type for the Virtual Machine",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"model": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"disk":     diskSchema(),
			"graphics": graphicsSchema(),
			"nic":      nicSchema(),
			"os":       osSchema(),
			"vmgroup":  vmGroupSchema(),
			"tags":     tagsSchema(),
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
			return nil, err
		}
		tc = controller.Template(int(gid))
	}

	// Otherwise, try to find the template by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.Templates().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
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
	}

	err = tc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	tplDef, err := generateTemplate(d)
	if err != nil {
		return err
	}

	tplID, err := controller.Templates().Create(tplDef)
	if err != nil {
		log.Printf("[ERROR] Template creation failed, error: %s", err)
		return err
	}

	tc := controller.Template(tplID)

	// add template information into Template
	template := d.Get("template").(string)
	if len(template) > 0 {
		err = tc.Update(template, 1)
		if err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprintf("%v", tplID))

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
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual machine template %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		return err
	}

	// TODO: fix it after 5.10 release availability
	// Force the "extended" bool to false to keep ONE 5.8 behavior
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	tpl, err := tc.Info(false, false)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", tpl.ID))
	d.Set("name", tpl.Name)
	d.Set("uid", tpl.UID)
	d.Set("gid", tpl.GID)
	d.Set("uname", tpl.UName)
	d.Set("gname", tpl.GName)
	d.Set("reg_time", tpl.RegTime)
	d.Set("permissions", permissionsUnixString(*tpl.Permissions))

	// Get Human readable tpl information
	tplstr := tpl.Template.String()

	err = d.Set("template", tplstr)
	if err != nil {
		return err
	}

	// Nics
	nics := tpl.Template.GetNICs()
	nicList := make([]interface{}, 0, len(nics))

	// Set Nics to resource
	for _, nic := range nics {
		nicList = append(nicList, flattenNIC(nic))
	}

	if len(nicList) > 0 {
		err = d.Set("nic", nicList)
		if err != nil {
			return err
		}
	}

	// Set Disks to resource
	disks := tpl.Template.GetDisks()
	diskList := make([]interface{}, 0, len(disks))

	// Set Disks to resource
	for _, disk := range disks {
		diskList = append(diskList, flattenDisk(disk))
	}

	if len(diskList) > 0 {
		err = d.Set("disk", diskList)
		if err != nil {
			return err
		}
	}

	err = flattenTemplate(d, &tpl.Template, true)
	if err != nil {
		return err
	}

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
	// TODO: fix it after 5.10 release availability
	// Force the "extended" bool to false to keep ONE 5.8 behavior
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	tpl, err := tc.Info(false, false)
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		err := tc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
		// Update Name in internal struct
		// TODO: fix it after 5.10 release availability
		// Force the "extended" bool to false to keep ONE 5.8 behavior
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		tpl, err = tc.Info(false, false)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated name for tpl %s\n", tpl.Name)
	}

	if d.HasChange("template") && d.Get("tpl") != "" {
		// replace the whole template instead of merging it with the existing one
		err = tc.Update(d.Get("template").(string), 1)
		if err != nil {
			return err
		}

		log.Printf("[INFO] Successfully updated template template %s\n", tpl.Name)

	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = tc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated Template %s\n", tpl.Name)
	}

	if d.HasChange("group") {
		err = changeTemplateGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for Template %s\n", tpl.Name)
	}

	if d.HasChange("tags") {
		tagsInterface := d.Get("tags").(map[string]interface{})
		for k, v := range tagsInterface {
			tpl.Template.Del(strings.ToUpper(k))
			tpl.Template.AddPair(strings.ToUpper(k), v.(string))
		}

		err = tc.Update(tpl.Template.String(), 1)
		if err != nil {
			return err
		}
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

func generateTemplate(d *schema.ResourceData) (string, error) {
	name := d.Get("name").(string)

	tpl := vm.NewTemplate()

	tpl.Add(vmk.Name, name)

	//Generate CONTEXT definition
	context := d.Get("context").(map[string]interface{})
	log.Printf("Number of CONTEXT vars: %d", len(context))
	log.Printf("CONTEXT Map: %s", context)

	// Add new context elements to the template
	for key, value := range context {
		keyUp := strings.ToUpper(key)
		tpl.AddCtx(vmk.Context(keyUp), fmt.Sprint(value))
	}

	generateVMTemplate(d, tpl)

	tplStr := tpl.String()
	log.Printf("[INFO] Template definitions: %s", tplStr)

	return tplStr, nil
}
