package opennebula

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
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
			"cpu":      cpuSchema(),
			"vcpu":     vcpuSchema(),
			"memory":   memorySchema(),
			"context":  contextSchema(),
			"cpumodel": cpumodelSchema(),
			"disk":     diskSchema(),
			"graphics": graphicsSchema(),
			"nic":      nicSchema(),
			"os":       osSchema(),
			"vmgroup":  vmGroupSchema(),
			"tags":     tagsSchema(),
			"raw": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Low-level hypervisor tuning",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								validtypes := []string{"kvm", "lxd", "vmware"}
								value := v.(string)

								if inArray(value, validtypes) < 0 {
									errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(validtypes, ",")))
								}

								return
							},
							Description: "Name of the hypervisor: kvm, lxd, vmware",
						},
						"data": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Low-level data to pass to the hypervisor",
						},
					},
				},
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
			"lock":                  lockSchema(),
			"sched_requirements":    schedReqSchema(),
			"sched_ds_requirements": schedDSReqSchema(),
			"description":           descriptionSchema(),
			"user_inputs": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Provides the template creator with the possibility to dynamically ask the user instantiating the template for dynamic values that must be defined.",
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
		group := d.Get("group").(string)
		gid, err = controller.Groups().ByName(group)
		if err != nil {
			return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
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

	if lock, ok := d.GetOk("lock"); ok && lock.(string) != "UNLOCK" {

		var level shared.LockLevel
		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			return err
		}

		err = tc.Lock(level)
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

	err = flattenTemplateDisks(d, &tpl.Template)
	if err != nil {
		return err
	}

	err = flattenTemplateNICs(d, &tpl.Template)
	if err != nil {
		return err
	}

	uInputs, _ := tpl.Template.GetVector("USER_INPUTS")
	if uInputs != nil && len(uInputs.Pairs) > 0 {
		uInputsMap := make(map[string]interface{}, 0)

		for _, ui := range uInputs.Pairs {
			uInputsMap[ui.Key()] = ui.Value
		}

		err = d.Set("user_inputs", uInputsMap)
		if err != nil {
			return err
		}
	}

	err = flattenTemplate(d, &tpl.Template)
	if err != nil {
		return err
	}

	err = flattenUserTemplate(d, &tpl.Template.Template)
	if err != nil {
		return err
	}

	rawVec, _ := tpl.Template.GetVector("RAW")
	if rawVec != nil {

		rawMap := make([]map[string]interface{}, 0, 1)

		hypType, _ := rawVec.GetStr("TYPE")
		data, _ := rawVec.GetStr("DATA")

		rawMap = append(rawMap, map[string]interface{}{
			"type": hypType,
			"data": data,
		})

		if _, ok := d.GetOk("raw"); ok {
			err = d.Set("raw", rawMap)
			if err != nil {
				return err
			}
		}
	}

	if tpl.LockInfos != nil {
		d.Set("lock", LockLevelToString(tpl.LockInfos.Locked))
	}

	return nil
}

func flattenTemplateNICs(d *schema.ResourceData, tpl *vm.Template) error {

	nics := tpl.GetNICs()
	nicList := make([]interface{}, 0, len(nics))

	for _, nic := range nics {
		nicList = append(nicList, flattenNIC(nic))
	}

	if len(nicList) > 0 {
		err := d.Set("nic", nicList)
		if err != nil {
			return err
		}
	}

	return nil
}

func flattenTemplateDisks(d *schema.ResourceData, tpl *vm.Template) error {

	disks := tpl.GetDisks()
	diskList := make([]interface{}, 0, len(disks))

	for _, disk := range disks {
		diskList = append(diskList, flattenDisk(disk))
	}

	if len(diskList) > 0 {
		err := d.Set("disk", diskList)
		if err != nil {
			return err
		}
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

	lock, lockOk := d.GetOk("lock")
	if d.HasChange("lock") && lockOk && lock.(string) == "UNLOCK" {

		err = tc.Unlock()
		if err != nil {
			return err
		}

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

	update := false
	newTpl := vm.NewTemplate()

	// copy existing template and re-escape chars if needed
	for _, e := range tpl.Template.Elements {
		pair, ok := e.(*dyn.Pair)
		if ok {
			// copy a pair
			escapedValue := strings.ReplaceAll(pair.Value, "\"", "\\\"")
			newTpl.AddPair(e.Key(), escapedValue)
		} else {
			// copy a vector
			vector, _ := e.(*dyn.Vector)

			newVec := newTpl.AddVector(e.Key())
			for _, p := range vector.Pairs {
				escapedValue := strings.ReplaceAll(p.Value, "\"", "\\\"")
				newVec.AddPair(p.Key(), escapedValue)
			}
		}
	}

	if d.HasChange("raw") {
		newTpl.Del("RAW")

		raw := d.Get("raw").([]interface{})
		if len(raw) > 0 {
			for i := 0; i < len(raw); i++ {
				rawConfig := raw[i].(map[string]interface{})
				rawVec := newTpl.AddVector("RAW")
				rawVec.AddPair("TYPE", rawConfig["type"].(string))
				rawVec.AddPair("DATA", rawConfig["data"].(string))
			}
		}
	}

	if d.HasChange("sched_requirements") {
		schedRequirements := d.Get("sched_requirements").(string)

		if len(schedRequirements) > 0 {
			escapedValue := strings.ReplaceAll(schedRequirements, "\"", "\\\"")
			newTpl.Placement(vmk.SchedRequirements, escapedValue)
		} else {
			newTpl.Del(string(vmk.SchedRequirements))
		}
		update = true
	}

	if d.HasChange("sched_ds_requirements") {
		schedDSRequirements := d.Get("sched_ds_requirements").(string)

		if len(schedDSRequirements) > 0 {
			escapedValue := strings.ReplaceAll(schedDSRequirements, "\"", "\\\"")
			newTpl.Placement(vmk.SchedDSRequirements, escapedValue)
		} else {
			newTpl.Del(string(vmk.SchedDSRequirements))
		}
		update = true
	}

	if d.HasChange("description") {
		newTpl.Del(string(vmk.Description))

		description := d.Get("description").(string)

		if len(description) > 0 {
			newTpl.Add(vmk.Description, description)
		}

		update = true
	}

	if d.HasChange("user_inputs") {
		newTpl.Del("USER_INPUTS")

		uInputs := d.Get("user_inputs").(map[string]interface{})
		if len(uInputs) > 0 {

			vec := newTpl.AddVector("USER_INPUTS")
			if len(uInputs) > 0 {
				for k, v := range uInputs {
					vec.AddPair(k, v)
				}
			}
		}

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
		err = tc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			return err
		}
	}

	if d.HasChange("lock") && lockOk && lock.(string) != "UNLOCK" {

		var level shared.LockLevel

		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			return err
		}

		err = tc.Lock(level)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaTemplateRead(d, meta)
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

	uInputs := d.Get("user_inputs").(map[string]interface{})
	if len(uInputs) > 0 {
		vec := tpl.AddVector("USER_INPUTS")

		for k, v := range uInputs {
			vec.AddPair(k, v)
		}
	}

	err := generateVMTemplate(d, tpl)
	if err != nil {
		return "", err
	}

	//Generate RAW definition
	raw := d.Get("raw").([]interface{})
	for i := 0; i < len(raw); i++ {
		rawConfig := raw[i].(map[string]interface{})
		rawVec := tpl.AddVector("RAW")
		rawVec.AddPair("TYPE", rawConfig["type"].(string))
		rawVec.AddPair("DATA", rawConfig["data"].(string))
	}

	tplStr := tpl.String()
	log.Printf("[INFO] Template definitions: %s", tplStr)

	return tplStr, nil
}
