package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	hookSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/hook"
	hookKey "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/hook/keys"
)

var hookTypes = []string{"api", "state"}
var hookResources = []string{"IMAGE", "HOST", "VM"}
var hookVMOns = []string{"PROLOG", "RUNNING", "SHUTDOWN", "STOP", "DONE", "UNKNOWN", "CUSTOM"}

var hookReservedTemplateKeys = map[string]struct{}{
	"NAME":            {},
	"TYPE":            {},
	"COMMAND":         {},
	"ARGUMENTS":       {},
	"ARGUMENTS_STDIN": {},
	"TIMEOUT":         {},
	"CALL":            {},
	"RESOURCE":        {},
	"REMOTE":          {},
	"STATE":           {},
	"LCM_STATE":       {},
	"ON":              {},
}

var hookShortcuts = map[string]struct {
	State    string
	LCMState string
}{
	"PROLOG":   {State: "ACTIVE", LCMState: "PROLOG"},
	"RUNNING":  {State: "ACTIVE", LCMState: "RUNNING"},
	"SHUTDOWN": {State: "ACTIVE", LCMState: "EPILOG"},
	"STOP":     {State: "STOPPED", LCMState: "LCM_INIT"},
	"DONE":     {State: "DONE", LCMState: "LCM_INIT"},
	"UNKNOWN":  {State: "ACTIVE", LCMState: "UNKNOWN"},
}

func resourceOpennebulaHook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaHookCreate,
		ReadContext:   resourceOpennebulaHookRead,
		UpdateContext: resourceOpennebulaHookUpdate,
		DeleteContext: resourceOpennebulaHookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customizeHookDiff,
		Schema:        hookSchema(false),
	}
}

func hookSchema(dataSource bool) map[string]*schema.Schema {
	idSchema := &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     -1,
		Description: "Id of the hook",
	}

	nameSchema := &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the hook",
	}
	typeSchema := &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Type of the hook: api or state",
		ValidateFunc: validation.StringInSlice(hookTypes, true),
		ForceNew:     true,
	}
	commandSchema := &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Command executed when the hook is triggered. Relative paths are resolved by OpenNebula's hook execution manager",
	}

	if dataSource {
		nameSchema.Required = false
		nameSchema.Optional = true
		typeSchema.Required = false
		typeSchema.Computed = true
		commandSchema.Required = false
		commandSchema.Computed = true
	} else {
		idSchema = nil
	}

	s := map[string]*schema.Schema{
		"name":    nameSchema,
		"type":    typeSchema,
		"command": commandSchema,
		"arguments": {
			Type:        schema.TypeString,
			Optional:    !dataSource,
			Computed:    dataSource,
			Description: "Arguments passed to the hook command",
		},
		"arguments_stdin": {
			Type:        schema.TypeBool,
			Optional:    !dataSource,
			Computed:    dataSource,
			Default:     boolDefault(dataSource, false),
			Description: "Pass arguments through stdin instead of command line arguments",
		},
		"timeout": {
			Type:        schema.TypeInt,
			Optional:    !dataSource,
			Computed:    dataSource,
			Description: "Hook execution timeout in seconds. If not set, OpenNebula uses an infinite timeout",
		},
		"call": {
			Type:        schema.TypeString,
			Optional:    !dataSource,
			Computed:    dataSource,
			Description: "API call that triggers an api hook, for example one.user.allocate",
		},
		"resource": {
			Type:         schema.TypeString,
			Optional:     !dataSource,
			Computed:     dataSource,
			Description:  "Resource type for state hooks: IMAGE, HOST, or VM",
			ValidateFunc: validation.StringInSlice(append([]string{""}, hookResources...), true),
		},
		"remote": {
			Type:        schema.TypeBool,
			Optional:    !dataSource,
			Computed:    dataSource,
			Default:     boolDefault(dataSource, false),
			Description: "Execute a state hook remotely on the host that triggered the hook or runs the VM",
		},
		"state": {
			Type:        schema.TypeString,
			Optional:    !dataSource,
			Computed:    true,
			Description: "State that triggers a state hook",
		},
		"lcm_state": {
			Type:        schema.TypeString,
			Optional:    !dataSource,
			Computed:    true,
			Description: "LCM state that triggers a VM state hook when on is CUSTOM",
		},
		"on": {
			Type:         schema.TypeString,
			Optional:     !dataSource,
			Computed:     dataSource,
			Description:  "Shortcut for common VM state hook transitions: PROLOG, RUNNING, SHUTDOWN, STOP, DONE, UNKNOWN, CUSTOM",
			ValidateFunc: validation.StringInSlice(append([]string{""}, hookVMOns...), true),
		},
		"tags":         tagsSchema(),
		"default_tags": defaultTagsSchemaComputed(),
		"tags_all":     tagsSchemaComputed(),
	}

	if dataSource {
		s["tags"].Computed = true

		for _, field := range s {
			if field.Computed && !field.Optional && !field.Required {
				field.ValidateDiagFunc = nil
				field.ValidateFunc = nil
			}
		}
	}

	if idSchema != nil {
		s["id"] = idSchema
	}

	return s
}

func boolDefault(dataSource bool, value bool) interface{} {
	if dataSource {
		return nil
	}
	return value
}

func customizeHookDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if err := SetTagsDiff(ctx, diff, meta); err != nil {
		return err
	}

	for k := range diff.Get("tags").(map[string]interface{}) {
		if _, reserved := hookReservedTemplateKeys[strings.ToUpper(k)]; reserved {
			return fmt.Errorf("tags cannot use hook template attribute %q", k)
		}
	}

	config := meta.(*Configuration)
	for k := range config.defaultTags {
		if _, reserved := hookReservedTemplateKeys[strings.ToUpper(k)]; reserved {
			return fmt.Errorf("default_tags cannot use hook template attribute %q for opennebula_hook", k)
		}
	}

	hookType := strings.ToLower(diff.Get("type").(string))
	switch hookType {
	case "api":
		if diff.Get("call").(string) == "" {
			return fmt.Errorf("call is required when type is api")
		}
		for _, attr := range []string{"resource", "state", "lcm_state", "on"} {
			if diff.Get(attr).(string) != "" {
				return fmt.Errorf("%s cannot be set when type is api", attr)
			}
		}
		if diff.Get("remote").(bool) {
			return fmt.Errorf("remote cannot be set when type is api")
		}
	case "state":
		if diff.Get("call").(string) != "" {
			return fmt.Errorf("call cannot be set when type is state")
		}
		resource := strings.ToUpper(diff.Get("resource").(string))
		if resource == "" {
			return fmt.Errorf("resource is required when type is state")
		}
		switch resource {
		case "VM":

			if !diff.NewValueKnown("on") { // Could be derived from previous expression
				return nil
			}

			on := strings.ToUpper(diff.Get("on").(string))

			rawConfig := diff.GetRawConfig()
			if rawConfig.IsNull() || !rawConfig.IsKnown() {
				return nil
			}

			if on == "CUSTOM" {
				for _, attr := range []string{"state", "lcm_state"} {
					value := rawConfig.GetAttr(attr)

					if value.IsNull() {
						return fmt.Errorf("%s is required when on is CUSTOM", attr)
					}

					if value.IsKnown() && value.AsString() == "" {
						return fmt.Errorf("%s cannot be empty when on is CUSTOM", attr)

					}
				}

			} else {

				for _, attr := range []string{"state", "lcm_state"} {
					if !rawConfig.GetAttr(attr).IsNull() {
						return fmt.Errorf(
							"%s cannot be configured when on is %q; use on = CUSTOM to specify states",
							attr, on,
						)
					}
				}
				shortcut, ok := hookShortcuts[on]

				if !ok {
					return fmt.Errorf("unsupported VM hook shortcut %q", on)

				}

				if err := diff.SetNew("state", shortcut.State); err != nil {
					return fmt.Errorf("error setting hook state: %w", err)
				}

				if err := diff.SetNew("lcm_state", shortcut.LCMState); err != nil {
					return fmt.Errorf("error setting hook lcm_state: %w", err)
				}
			}

		case "HOST", "IMAGE":
			if diff.Get("state").(string) == "" {
				return fmt.Errorf("state is required for %s state hooks", resource)
			}
			if diff.Get("on").(string) != "" {
				return fmt.Errorf("on can only be set for VM state hooks")
			}
			if diff.Get("lcm_state").(string) != "" {
				return fmt.Errorf("lcm_state can only be set for VM state hooks")
			}
		}
	}

	return nil
}

func getHookController(d *schema.ResourceData, meta interface{}) (*goca.HookController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	hookID, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.Hook(int(hookID)), nil
}

func resourceOpennebulaHookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	tpl, err := generateHookTemplate(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Hook template: %s\n", tpl.String())
	hookID, err := controller.Hooks().CreateContext(ctx, tpl.String())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the hook",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%v", hookID))

	return resourceOpennebulaHookRead(ctx, d, meta)
}

func resourceOpennebulaHookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	hc, err := getHookController(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	hook, err := hc.InfoContext(ctx, false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing hook %s from state because it no longer exists", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve hook information",
			Detail:   fmt.Sprintf("hook (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	diags = append(diags, setHookResourceData(d, meta, hook)...)
	if diags.HasError() {
		return diags
	}

	return nil
}

func resourceOpennebulaHookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	hc, err := getHookController(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") {
		err = hc.RenameContext(ctx, d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename hook",
				Detail:   fmt.Sprintf("hook (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	tpl, err := generateHookTemplate(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	err = hc.UpdateContext(ctx, tpl.String(), parameters.Replace)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to update hook content",
			Detail:   fmt.Sprintf("hook (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return resourceOpennebulaHookRead(ctx, d, meta)
}

func resourceOpennebulaHookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	hc, err := getHookController(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	err = hc.DeleteContext(ctx)
	if err != nil && !NoExists(err) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete hook",
			Detail:   fmt.Sprintf("hook (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully deleted Hook ID %s\n", d.Id())
	return nil
}

func generateHookTemplate(d *schema.ResourceData, meta interface{}) (*hookSc.Template, error) {
	config := meta.(*Configuration)

	tpl := hookSc.NewTemplate()
	tpl.Add(hookKey.Name, d.Get("name").(string))
	tpl.Add(hookKey.Type, strings.ToLower(d.Get("type").(string)))
	tpl.Add(hookKey.Command, d.Get("command").(string))

	addHookOptionalString(tpl, "ARGUMENTS", d.Get("arguments").(string))
	addHookOptionalYesNo(tpl, "ARGUMENTS_STDIN", d.Get("arguments_stdin").(bool))
	addHookOptionalInt(tpl, "TIMEOUT", d.Get("timeout").(int))

	switch strings.ToLower(d.Get("type").(string)) {
	case "api":
		addHookOptionalString(tpl, "CALL", d.Get("call").(string))
	case "state":
		addHookOptionalString(tpl, "RESOURCE", strings.ToUpper(d.Get("resource").(string)))
		addHookOptionalYesNo(tpl, "REMOTE", d.Get("remote").(bool))
		addHookOptionalString(tpl, "STATE", strings.ToUpper(d.Get("state").(string)))
		addHookOptionalString(tpl, "LCM_STATE", strings.ToUpper(d.Get("lcm_state").(string)))
		addHookOptionalString(tpl, "ON", strings.ToUpper(d.Get("on").(string)))
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		key := strings.ToUpper(k)
		if _, reserved := hookReservedTemplateKeys[key]; reserved {
			return nil, fmt.Errorf("tags cannot use hook template attribute %q", k)
		}
		tpl.AddPair(key, v)
	}

	if len(config.defaultTags) > 0 {
		for k, v := range config.defaultTags {
			key := strings.ToUpper(k)
			if _, reserved := hookReservedTemplateKeys[key]; reserved {
				return nil, fmt.Errorf("default_tags cannot use hook template attribute %q for opennebula_hook", k)
			}
			p, _ := tpl.GetPair(key)
			if p != nil {
				continue
			}
			tpl.AddPair(key, v)
		}
	}

	return tpl, nil
}

func addHookOptionalString(tpl *hookSc.Template, key string, value string) {
	if value != "" {
		tpl.AddPair(key, value)
	}
}

func addHookOptionalInt(tpl *hookSc.Template, key string, value int) {
	if value != 0 {
		tpl.AddPair(key, value)
	}
}

func addHookOptionalYesNo(tpl *hookSc.Template, key string, value bool) {
	if value {
		tpl.AddPair(key, "YES")
	}
}

func setHookResourceData(d *schema.ResourceData, meta interface{}, hook *hookSc.Hook) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Set("name", hook.Name)
	hookType := strings.ToLower(hook.Type)
	if hookType == "" {
		hookType = strings.ToLower(getHookTemplateString(&hook.Template.Template, "TYPE"))
	}
	d.Set("type", hookType)

	setHookString(d, &hook.Template.Template, "command", "COMMAND")
	setHookString(d, &hook.Template.Template, "arguments", "ARGUMENTS")
	d.Set("arguments_stdin", parseHookBool(getHookTemplateString(&hook.Template.Template, "ARGUMENTS_STDIN")))
	setHookInt(d, &hook.Template.Template, "timeout", "TIMEOUT")
	setHookString(d, &hook.Template.Template, "call", "CALL")
	setHookUpperString(d, &hook.Template.Template, "resource", "RESOURCE")
	d.Set("remote", parseHookBool(getHookTemplateString(&hook.Template.Template, "REMOTE")))
	setHookUpperString(d, &hook.Template.Template, "state", "STATE")
	setHookUpperString(d, &hook.Template.Template, "lcm_state", "LCM_STATE")

	if hookType == "state" && d.Get("resource").(string) == "VM" {
		if d.Get("on").(string) == "" {
			d.Set("on", "CUSTOM")
		}
	} else {
		d.Set("on", "")
	}

	diags = append(diags, flattenTemplateTags(d, meta, &hook.Template.Template)...)
	return diags
}

func getHookTemplateString(tpl *dynamic.Template, key string) string {
	value, err := tpl.GetStr(key)
	if err != nil {
		return ""
	}
	return value
}

func setHookString(d *schema.ResourceData, tpl *dynamic.Template, attr string, key string) {
	d.Set(attr, getHookTemplateString(tpl, key))
}

func setHookUpperString(d *schema.ResourceData, tpl *dynamic.Template, attr string, key string) {
	d.Set(attr, strings.ToUpper(getHookTemplateString(tpl, key)))
}

func setHookInt(d *schema.ResourceData, tpl *dynamic.Template, attr string, key string) {
	value := getHookTemplateString(tpl, key)
	if value == "" {
		return
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("[WARN] Unable to parse hook %s value %q as integer: %s", key, value, err)
		return
	}
	d.Set(attr, i)
}

func parseHookBool(value string) bool {
	switch strings.ToLower(value) {
	case "yes", "true", "1":
		return true
	default:
		return false
	}
}

func hookTemplateTags(tpl dynamic.Template) map[string]interface{} {
	tags := make(map[string]interface{})

	for i := range tpl.Elements {
		pair, ok := tpl.Elements[i].(*dynamic.Pair)
		if !ok {
			continue
		}

		if _, reserved := hookReservedTemplateKeys[strings.ToUpper(pair.Key())]; reserved {
			continue
		}

		tags[pair.Key()] = pair.Value
	}

	return tags
}
