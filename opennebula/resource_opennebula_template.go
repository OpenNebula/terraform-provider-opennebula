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
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/template"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

func resourceOpennebulaTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaTemplateCreate,
		ReadContext:   resourceOpennebulaTemplateRead,
		Exists:        resourceOpennebulaTemplateExists,
		UpdateContext: resourceOpennebulaTemplateUpdate,
		DeleteContext: resourceOpennebulaTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: SetTagsDiff,
		Schema: mergeSchemas(
			commonTemplateSchemas(),
			map[string]*schema.Schema{
				"nic": nicSchema(),
			},
		),
	}
}

func commonTemplateSchemas() map[string]*schema.Schema {
	return mergeSchemas(
		commonInstanceSchema(),
		map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the template",
			},
			"features": {
				Type:        schema.TypeSet,
				Optional:    true,
				MinItems:    1,
				Description: "List of Features",
				Elem: &schema.Resource{
					Schema: FeaturesFields(),
				},
			},
			"disk": diskSchema(),
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
			"reg_time": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Registration time",
			},
			"user_inputs": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Provides the template creator with the possibility to dynamically ask the user instantiating the template for dynamic values that must be defined.",
			},
		},
	)
}

func FeaturesFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"pae": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Physical address extension mode allows 32-bit guests to address more than 4 GB of memory.",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				validtypes := []string{"YES", "NO"}
				value := v.(string)

				if inArray(value, validtypes) < 0 {
					errors = append(errors, fmt.Errorf("PAE must be one of: %s", strings.Join(validtypes, ", ")))
				}

				return
			},
		},
		"acpi": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Useful for power management, for example, with KVM guests it is required for graceful shutdown to work.",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				validtypes := []string{"YES", "NO"}
				value := v.(string)

				if inArray(value, validtypes) < 0 {
					errors = append(errors, fmt.Errorf("ACPI must be one of: %s", strings.Join(validtypes, ", ")))
				}

				return
			},
		},
		"apic": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Enables the advanced programmable IRQ management. Useful for SMP machines.",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				validtypes := []string{"YES", "NO"}
				value := v.(string)

				if inArray(value, validtypes) < 0 {
					errors = append(errors, fmt.Errorf("APIC must be one of: %s", strings.Join(validtypes, ", ")))
				}

				return
			},
		},
		"localtime": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The guest clock will be synchronized to the host’s configured timezone when booted. Useful for Windows VMs.",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				validtypes := []string{"YES", "NO"}
				value := v.(string)

				if inArray(value, validtypes) < 0 {
					errors = append(errors, fmt.Errorf("LOCALTIME must be one of: %s", strings.Join(validtypes, ", ")))
				}

				return
			},
		},
		"hyperv": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Add hyperv extensions to the VM. The options can be configured in the driver configuration, HYPERV_OPTIONS.",
		},
		"guest_agent": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Enables the QEMU Guest Agent communication. This only creates the socket inside the VM, the Guest Agent itself must be installed and started in the VM.",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				validtypes := []string{"YES", "NO"}
				value := v.(string)

				if inArray(value, validtypes) < 0 {
					errors = append(errors, fmt.Errorf("GUEST_AGENT must be one of: %s", strings.Join(validtypes, ", ")))
				}

				return
			},
		},
		"virtio_scsi_queues": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Numer of vCPU queues for the virtio-scsi controller.",
		},
		"iothreads": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Number of iothreads for virtio disks. By default threads will be assign to disk by round robin algorithm. Disk thread id can be forced by disk IOTHREAD attribute.",
		},
	}
}

func getTemplateController(d *schema.ResourceData, meta interface{}) (*goca.TemplateController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	tplID, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.Template(int(tplID)), nil
}

func changeTemplateGroup(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller
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

func resourceOpennebulaTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diags := resourceOpennebulaTemplateCreateCustom(ctx, d, meta, func(d *schema.ResourceData, tpl *dyn.Template) diag.Diagnostics {

		//Generate NIC definition
		nics := d.Get("nic").([]interface{})
		log.Printf("Number of NICs: %d", len(nics))

		for i := 0; i < len(nics); i++ {
			nicconfig := nics[i].(map[string]interface{})

			nic := makeNICVector(nicconfig)
			tpl.Elements = append(tpl.Elements, nic)
		}

		return nil
	})

	if len(diags) > 0 {
		return diags
	}

	return resourceOpennebulaTemplateRead(ctx, d, meta)
}

func resourceOpennebulaTemplateCreateCustom(ctx context.Context, d *schema.ResourceData, meta interface{}, customFunc customDynTemplateFunc) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	tpl, err := generateTemplate(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate description",
			Detail:   err.Error(),
		})
		return diags
	}

	if customFunc != nil {
		customDiags := customFunc(d, &tpl.Template)
		if len(customDiags) > 0 {
			return customDiags
		}
	}

	tplDef := tpl.Template.String()
	log.Printf("[INFO] Template definitions: %s", tplDef)
	tplID, err := controller.Templates().Create(tplDef)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the template",
			Detail:   err.Error(),
		})
		return diags
	}

	tc := controller.Template(tplID)

	d.SetId(fmt.Sprintf("%v", tplID))

	// Change Permissions only if Permissions are set
	if perms, ok := d.GetOk("permissions"); ok {
		err = tc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err = changeTemplateGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
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
				Summary:  "Failed to convert lock level",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = tc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return nil
}

func resourceOpennebulaTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceOpennebulaTemplateReadCustom(ctx, d, meta, templateReadCustom)
}

func templateReadCustom(ctx context.Context, d *schema.ResourceData, templateInfos *template.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	err := flattenTemplateNICs(d, &templateInfos.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten NICs",
			Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return nil
}

func resourceOpennebulaTemplateReadCustom(ctx context.Context, d *schema.ResourceData, meta interface{}, readCustom customTemplateFunc) diag.Diagnostics {

	var diags diag.Diagnostics

	// Get requested template from all templates
	tc, err := getTemplateController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the template controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// TODO: fix it after 5.10 release availability
	// Force the "extended" bool to false to keep ONE 5.8 behavior
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	tpl, err := tc.Info(false, false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual machine template %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", tpl.ID))
	d.Set("name", tpl.Name)
	d.Set("uid", tpl.UID)
	d.Set("gid", tpl.GID)
	d.Set("uname", tpl.UName)
	d.Set("gname", tpl.GName)
	d.Set("reg_time", tpl.RegTime)
	d.Set("permissions", permissionsUnixString(*tpl.Permissions))

	err = flattenTemplateDisks(d, &tpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten template disks",
			Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	uInputs, _ := tpl.Template.GetVector("USER_INPUTS")
	if uInputs != nil && len(uInputs.Pairs) > 0 {
		uInputsMap := make(map[string]interface{}, 0)

		for _, ui := range uInputs.Pairs {
			uInputsMap[ui.Key()] = ui.Value
		}

		err = d.Set("user_inputs", uInputsMap)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to set attribute",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if readCustom != nil {
		customDiags := readCustom(ctx, d, tpl)
		if len(customDiags) > 0 {
			return customDiags
		}
	}

	err = flattenTemplate(d, nil, &tpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten template",
			Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	err = flattenFeatures(d, &tpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten features",
			Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	flattenDiags := flattenVMUserTemplate(d, meta, nil, &tpl.Template.Template)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("template (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
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
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to set attribute",
					Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
	}

	if tpl.LockInfos != nil {
		d.Set("lock", LockLevelToString(tpl.LockInfos.Locked))
	}

	return nil
}

func flattenFeatures(d *schema.ResourceData, tpl *vm.Template) error {
	// Features
	featuresMap := make([]map[string]interface{}, 0, 1)
	pae, _ := tpl.GetFeature(vmk.PAE)
	acpi, _ := tpl.GetFeature(vmk.ACPI)
	apic, _ := tpl.GetFeature(vmk.APIC)
	localtime, _ := tpl.GetFeature(vmk.LocalTime)
	hyperv, _ := tpl.GetFeature("HYPERV")
	guest_agent, _ := tpl.GetFeature(vmk.GuestAgent)
	virtio_scsi_queues, _ := tpl.GetFeature(vmk.VirtIOScsiQueues)
	iothreads, _ := tpl.GetFeature("IOTHREADS")

	// Set features to resource
	if pae != "" || acpi != "" || apic != "" || localtime != "" || hyperv != "" || guest_agent != "" || virtio_scsi_queues != "" || iothreads != "" {
		featuresMap = append(featuresMap, map[string]interface{}{
			"pae":                pae,
			"acpi":               acpi,
			"apic":               apic,
			"localtime":          localtime,
			"hyperv":             hyperv,
			"guest_agent":        guest_agent,
			"virtio_scsi_queues": virtio_scsi_queues,
			"iothreads":          iothreads,
		})

		err := d.Set("features", featuresMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func flattenTemplateNICs(d *schema.ResourceData, tpl *vm.Template) error {

	nics := tpl.GetNICs()
	nicList := make([]interface{}, 0, len(nics))

	for _, nic := range nics {
		nicList = append(nicList, flattenNIC(nic))
	}

	err := d.Set("nic", nicList)
	if err != nil {
		return err
	}

	return nil
}

func flattenTemplateDisks(d *schema.ResourceData, tpl *vm.Template) error {

	disks := tpl.GetDisks()
	diskList := make([]interface{}, 0, len(disks))

	for _, disk := range disks {
		diskList = append(diskList, flattenDisk(disk))
	}

	err := d.Set("disk", diskList)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaTemplateExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	serviceTemplateID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.Template(int(serviceTemplateID)).Info(false, false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaTemplateUpdateCustom(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	//Get Template
	tc, err := getTemplateController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the template controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// TODO: fix it after 5.10 release availability
	// Force the "extended" bool to false to keep ONE 5.8 behavior
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	tpl, err := tc.Info(false, false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	lock, lockOk := d.GetOk("lock")
	if d.HasChange("lock") && lockOk && lock.(string) == "UNLOCK" {

		err = tc.Unlock()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to unlock",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("name") {
		err := tc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		// Update Name in internal struct
		// TODO: fix it after 5.10 release availability
		// Force the "extended" bool to false to keep ONE 5.8 behavior
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		tpl, err = tc.Info(false, false)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve informations",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name for tpl %s\n", tpl.Name)
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = tc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change permissions",
					Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated Template %s\n", tpl.Name)
	}

	if d.HasChange("group") {
		err = changeTemplateGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated group for Template %s\n", tpl.Name)
	}

	update := false
	newTpl := tpl.Template

	if d.HasChange("features") {
		newTpl.Del("FEATURES")

		features := d.Get("features").(*schema.Set).List()

		for _, featuresInterface := range features {
			featuresMap := featuresInterface.(map[string]interface{})
			log.Printf("Number of FEATURES vars: %d", len(featuresMap))
			log.Printf("FEATURES Map: %s", featuresMap)
			for key, value := range featuresMap {
				if value != "" {
					keyUp := strings.ToUpper(key)
					newTpl.AddFeature(vmk.Feature(keyUp), fmt.Sprint(value))
				}
			}
		}

		update = true
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
			newTpl.Placement(vmk.SchedRequirements, schedRequirements)
		} else {
			newTpl.Del(string(vmk.SchedRequirements))
		}
		update = true
	}

	if d.HasChange("sched_ds_requirements") {
		schedDSRequirements := d.Get("sched_ds_requirements").(string)

		if len(schedDSRequirements) > 0 {
			newTpl.Placement(vmk.SchedDSRequirements, schedDSRequirements)
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

	if d.HasChange("cpumodel") {
		newTpl.Del("CPU_MODEL")
		cpumodel := d.Get("cpumodel").([]interface{})

		for i := 0; i < len(cpumodel); i++ {
			cpumodelconfig := cpumodel[i].(map[string]interface{})
			newTpl.CPUModel(cpumodelconfig["model"].(string))
		}

	}

	if d.HasChange("cpu") {
		cpu := d.Get("cpu").(float64)
		if cpu > 0 {
			update = true
			newTpl.CPU(cpu)
		}
	}

	if d.HasChange("vcpu") {
		vcpu := d.Get("vcpu").(int)
		if vcpu > 0 {
			update = true
			newTpl.VCPU(vcpu)
		}
	}

	if d.HasChange("memory") {
		memory := d.Get("memory").(int)
		if memory > 0 {
			update = true
			newTpl.Memory(memory)
		}
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
			key := strings.ToUpper(k)
			newTpl.Del(key)
			newTpl.AddPair(key, v)
		}

		update = true
	}

	if d.HasChange("tags_all") {
		oldTagsAllIf, newTagsAllIf := d.GetChange("tags_all")
		oldTagsAll := oldTagsAllIf.(map[string]interface{})
		newTagsAll := newTagsAllIf.(map[string]interface{})

		tags := d.Get("tags").(map[string]interface{})

		// delete tags
		for k, _ := range oldTagsAll {
			_, ok := newTagsAll[k]
			if ok {
				continue
			}
			newTpl.Del(strings.ToUpper(k))
		}

		// reapply all default tags that were neither applied nor overriden via tags section
		for k, v := range newTagsAll {
			_, ok := tags[k]
			if ok {
				continue
			}

			key := strings.ToUpper(k)
			newTpl.Del(key)
			newTpl.AddPair(key, v)
		}

		update = true
	}

	if update {
		err = tc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
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
				Summary:  "Failed to convert lock level",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = tc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return nil
}

func resourceOpennebulaTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	err := resourceOpennebulaTemplateUpdateCustom(ctx, d, meta)
	if err != nil {
		return nil
	}

	return resourceOpennebulaTemplateRead(ctx, d, meta)
}

func resourceOpennebulaTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	tc, err := getTemplateController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the template controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = tc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully deleted Template ID %s\n", d.Id())

	return nil
}

func generateTemplate(d *schema.ResourceData, meta interface{}) (*vm.Template, error) {
	name := d.Get("name").(string)

	tpl := vm.NewTemplate()

	tpl.Add(vmk.Name, name)

	//Generate FEATURES definition
	features := d.Get("features").(*schema.Set).List()

	for _, featuresInterface := range features {
		featuresMap := featuresInterface.(map[string]interface{})
		log.Printf("Number of FEATURES vars: %d", len(featuresMap))
		log.Printf("FEATURES Map: %s", featuresMap)
		for key, value := range featuresMap {
			if value != "" {
				keyUp := strings.ToUpper(key)
				tpl.AddFeature(vmk.Feature(keyUp), fmt.Sprint(value))
			}
		}
	}

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
		return nil, err
	}

	//Generate RAW definition
	raw := d.Get("raw").([]interface{})
	for i := 0; i < len(raw); i++ {
		rawConfig := raw[i].(map[string]interface{})
		rawVec := tpl.AddVector("RAW")
		rawVec.AddPair("TYPE", rawConfig["type"].(string))
		rawVec.AddPair("DATA", rawConfig["data"].(string))
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

	return tpl, nil
}
