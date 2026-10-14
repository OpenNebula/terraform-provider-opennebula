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
			map[string]*schema.Schema{
				"nic_alias": nicAliasSchema(),
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
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Feature flags for the VM. A single block; changed attributes are updated in place.",
				Elem: &schema.Resource{
					Schema: FeaturesFields(),
				},
			},
			"disk": diskSchema(),
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

// OpenNebula stores this FEATURES attribute as LOCALTIME.
// goca's vmk.LocalTime constant is LOCAL_TIME, which never matches.
const featureLocalTime vmk.Feature = "LOCALTIME"

// goca has no constant for this FEATURES attribute.
const featureVirtioBlkQueues vmk.Feature = "VIRTIO_BLK_QUEUES"

func FeaturesFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"pae": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Physical address extension mode allows 32-bit guests to address more than 4 GB of memory.",
			ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
				validtypes := []string{"YES", "NO"}
				value := v.(string)

				if !contains(value, validtypes) {
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

				if !contains(value, validtypes) {
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

				if !contains(value, validtypes) {
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

				if !contains(value, validtypes) {
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

				if !contains(value, validtypes) {
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
		"virtio_blk_queues": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Number of queues for the virtio-blk driver.",
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

		//Generate NIC ALIAS definition
		nicAliases, ok := d.Get("nic_alias").([]any)
		if !ok {
			return diag.FromErr(fmt.Errorf("invalid NIC Aliases configuration: %v", d.Get("nic_alias")))
		}
		log.Printf("Number of NIC Aliases: %d", len(nicAliases))

		for _, nicAlias := range nicAliases {
			nicAliasconfig, ok := nicAlias.(map[string]any)
			if !ok {
				return diag.FromErr(fmt.Errorf("invalid NIC Alias configuration: %v", nicAlias))
			}

			nicAlias := makeNICAliasVector(nicAliasconfig)
			tpl.Elements = append(tpl.Elements, nicAlias)
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

	if d.Get("group") != "" {
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

	err = flattenTemplateNICAliases(d, &templateInfos.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten NIC Aliases",
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

	err = flattenContext(d, &tpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten context",
			Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	flattenDiags := flattenVMUserTemplate(d, meta, nil, &tpl.Template.Template)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("template (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
	}

	if tpl.LockInfos != nil {
		d.Set("lock", LockLevelToString(tpl.LockInfos.Locked))
	}

	return nil
}

func configuredFeatures(d *schema.ResourceData) []map[string]interface{} {
	raw, ok := d.Get("features").([]interface{})
	if !ok {
		return nil
	}

	features := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if item == nil {
			continue
		}
		features = append(features, item.(map[string]interface{}))
	}

	return features
}

// addFeatures writes FEATURES attributes onto an existing template.
// Empty values remove that attribute. Other FEATURES keys are left in place.
func addFeatures(tpl *vm.Template, features []map[string]interface{}) {
	if len(features) == 0 {
		tpl.Del(vmk.FeaturesVec)
		return
	}

	for _, featuresMap := range features {
		log.Printf("Number of FEATURES vars: %d", len(featuresMap))
		log.Printf("FEATURES Map: %s", featuresMap)
		for key, value := range featuresMap {
			keyUp := strings.ToUpper(key)
			if keyUp == string(featureLocalTime) {
				deleteFeature(tpl, string(vmk.LocalTime))
			}
			deleteFeature(tpl, keyUp)
			if value == nil || value == "" {
				continue
			}
			if err := tpl.AddFeature(vmk.Feature(keyUp), fmt.Sprint(value)); err != nil {
				log.Printf("[ERROR] Failed to add feature %s: %s", keyUp, err)
			}
		}
	}
}

func addContext(tpl *vm.Template, context map[string]interface{}) {
	tpl.Del(vmk.ContextVec)
	if len(context) == 0 {
		return
	}

	log.Printf("Number of CONTEXT vars: %d", len(context))
	log.Printf("CONTEXT Map: %s", context)
	for key, value := range context {
		keyUp := strings.ToUpper(key)
		if err := tpl.AddCtx(vmk.Context(keyUp), fmt.Sprint(value)); err != nil {
			log.Printf("[ERROR] Failed to add context %s: %s", keyUp, err)
		}
	}
}

func contextValueMatches(configured, remote string) bool {
	if normalizeRawData(configured) == normalizeRawData(remote) {
		return true
	}
	return normalizeRawData(configured) == normalizeRawData(unescapeTemplateValue(remote))
}

// decodedTemplateValue returns the text OpenNebula stored, without a single
// extra escape layer of \ and ".
func decodedTemplateValue(remote string) string {
	unescaped := unescapeTemplateValue(remote)
	if escapeTemplateValue(unescaped) == remote {
		return unescaped
	}
	return remote
}

// decodeContextValues removes the escape layer OpenNebula returns on CONTEXT
// attributes. Updates reserialize the whole template, and that serialization
// escapes \ and " again. Leaving the returned layer in place writes a second
// set of backslashes onto values such as START_SCRIPT.
func decodeContextValues(tpl *vm.Template) {
	for _, vec := range tpl.GetVectors(string(vmk.ContextVec)) {
		for i := range vec.Pairs {
			vec.Pairs[i].Value = decodedTemplateValue(vec.Pairs[i].Value)
		}
	}
}

func flattenContext(d *schema.ResourceData, tpl *vm.Template) error {
	if len(tpl.GetVectors(vmk.ContextVec)) == 0 {
		return d.Set("context", map[string]interface{}{})
	}

	vec, err := tpl.GetVector(vmk.ContextVec)
	if err != nil {
		return err
	}

	existing := map[string]interface{}{}
	if raw, ok := d.GetOk("context"); ok {
		existing = raw.(map[string]interface{})
	}

	stateKeyByUpper := make(map[string]string, len(existing))
	for key := range existing {
		stateKeyByUpper[strings.ToUpper(key)] = key
	}

	contextMap := make(map[string]interface{}, len(vec.Pairs))
	for _, pair := range vec.Pairs {
		key := pair.Key()
		if orig, ok := stateKeyByUpper[strings.ToUpper(key)]; ok {
			key = orig
			// Keep the configured text when OpenNebula only trimmed it or
			// stored an extra escape of quotes and backslashes.
			if current, ok := existing[key].(string); ok && contextValueMatches(current, pair.Value) {
				contextMap[key] = current
				continue
			}
		}
		contextMap[key] = decodedTemplateValue(pair.Value)
	}

	return d.Set("context", contextMap)
}

func deleteFeature(tpl *vm.Template, key string) {
	vectors := tpl.GetVectors(vmk.FeaturesVec)
	if len(vectors) == 1 {
		vectors[0].Del(key)
	}
}

func featureValue(tpl *vm.Template, key vmk.Feature) string {
	value, _ := tpl.GetFeature(key)
	return value
}

func flattenFeatures(d *schema.ResourceData, tpl *vm.Template) error {
	// Features
	featuresMap := make([]map[string]interface{}, 0, 1)
	pae := featureValue(tpl, vmk.PAE)
	acpi := featureValue(tpl, vmk.ACPI)
	apic := featureValue(tpl, vmk.APIC)
	localtime := featureValue(tpl, featureLocalTime)
	if localtime == "" {
		// Older templates may have been stored with goca's LOCAL_TIME spelling.
		localtime = featureValue(tpl, vmk.LocalTime)
	}
	hyperv := featureValue(tpl, vmk.HyperV)
	guest_agent := featureValue(tpl, vmk.GuestAgent)
	virtio_scsi_queues := featureValue(tpl, vmk.VirtIOScsiQueues)
	virtio_blk_queues := featureValue(tpl, featureVirtioBlkQueues)
	iothreads := featureValue(tpl, vmk.IOThreads)

	// Set features to resource
	if pae != "" || acpi != "" || apic != "" || localtime != "" || hyperv != "" || guest_agent != "" || virtio_scsi_queues != "" || virtio_blk_queues != "" || iothreads != "" {
		featuresMap = append(featuresMap, map[string]interface{}{
			"pae":                pae,
			"acpi":               acpi,
			"apic":               apic,
			"localtime":          localtime,
			"hyperv":             hyperv,
			"guest_agent":        guest_agent,
			"virtio_scsi_queues": virtio_scsi_queues,
			"virtio_blk_queues":  virtio_blk_queues,
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

func flattenTemplateNICAliases(d *schema.ResourceData, tpl *vm.Template) error {

	nicAliases := tpl.GetNICAliases()
	nicAliasList := make([]any, 0, len(nicAliases))

	for _, nicAlias := range nicAliases {
		nicAliasList = append(nicAliasList, flattenNICAlias(nicAlias))
	}

	err := d.Set("nic_alias", nicAliasList)
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
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
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
	// Decode before applying config. addContext then replaces CONTEXT with the
	// raw configured text when that block itself changed.
	decodeContextValues(&newTpl)

	if d.HasChange("features") {
		addFeatures(&newTpl, configuredFeatures(d))
		update = true
	}

	if d.HasChange("context") {
		addContext(&newTpl, d.Get("context").(map[string]interface{}))
		update = true
	}

	if d.HasChange("raw") {
		updateRaw(d, &newTpl.Template)
		update = true
	}

	if updateTemplate(d, &newTpl) {
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

	if d.HasChange("disk") {
		newTpl.Del("DISK")

		err := addDisks(d, &newTpl)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update disks",
				Detail:   fmt.Sprintf("template (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("nic") {
		newTpl.Del(shared.NICVec)
		addNICs(d, &newTpl)
	}

	if d.HasChange("nic_alias") {
		newTpl.Del(shared.NICAliasVec)
		addNICs(d, &newTpl)
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

	addFeatures(tpl, configuredFeatures(d))

	addContext(tpl, d.Get("context").(map[string]interface{}))

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
