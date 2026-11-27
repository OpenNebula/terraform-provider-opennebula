package opennebula

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var contextEntryNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func contextWriteOnlySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"version": {Type: schema.TypeInt, Required: true},
			"entry": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"name": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringMatch(contextEntryNamePattern, "use letters, digits, and underscores"),
					},
					"value": {Type: schema.TypeString, Required: true, Sensitive: true, WriteOnly: true},
				}},
			},
		}},
	}
}

// Only version and entry names are readable from state, including during refresh.
func contextWriteOnlyNames(value interface{}) map[string]bool {
	blocks, _ := value.([]interface{})
	if len(blocks) == 0 || blocks[0] == nil {
		return nil
	}
	entries, _ := blocks[0].(map[string]interface{})["entry"].([]interface{})
	names := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry != nil {
			name := entry.(map[string]interface{})["name"].(string)
			names[strings.ToUpper(name)] = true
		}
	}
	return names
}

func contextForWrite(d *schema.ResourceData) (map[string]interface{}, error) {
	context := d.Get("context").(map[string]interface{})
	if _, configured := d.GetOk("context_wo"); !configured {
		return context, nil
	}

	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil, fmt.Errorf("context_wo values are unavailable in the Terraform configuration")
	}
	blocks := config.GetAttr("context_wo")
	if blocks.IsNull() || !blocks.IsKnown() || blocks.LengthInt() != 1 {
		return nil, fmt.Errorf("context_wo block is unavailable in the Terraform configuration")
	}
	entries := blocks.AsValueSlice()[0].GetAttr("entry")
	if entries.IsNull() || !entries.IsKnown() {
		return nil, fmt.Errorf("context_wo entries are unavailable in the Terraform configuration")
	}

	seen := map[string]bool{}
	regular := config.GetAttr("context")
	if regular.IsKnown() && !regular.IsNull() {
		for name := range regular.AsValueMap() {
			seen[strings.ToUpper(name)] = true
		}
	}

	merged := make(map[string]interface{}, len(context)+entries.LengthInt())
	for name, value := range context {
		merged[strings.ToUpper(name)] = value
	}
	for it := entries.ElementIterator(); it.Next(); {
		_, entry := it.Element()
		name, value := entry.GetAttr("name"), entry.GetAttr("value")
		if name.IsNull() || !name.IsKnown() || value.IsNull() || !value.IsKnown() {
			return nil, fmt.Errorf("context_wo entry name or value is unavailable")
		}
		key := strings.ToUpper(name.AsString())
		if !contextEntryNamePattern.MatchString(key) {
			return nil, fmt.Errorf("context_wo entry names must use letters, digits, and underscores")
		}
		if seen[key] {
			return nil, fmt.Errorf("context_wo entry name %q must be unique across context and context_wo", key)
		}
		seen[key] = true
		merged[key] = value.AsString()
	}
	return merged, nil
}

func hasWriteOnlyContextChange(d *schema.ResourceData) bool {
	_, configured := d.GetOk("context_wo")
	return d.HasChange("context_wo") || (configured && d.HasChange("context"))
}

// Apply managed entries to the remote vector while retaining inherited entries.
func updateWriteOnlyContext(d *schema.ResourceData, tpl *vm.Template) error {
	context, err := contextForWrite(d)
	if err != nil {
		return err
	}
	if len(tpl.GetVectors(vmk.ContextVec)) == 0 {
		addContext(tpl, context)
		return nil
	}
	oldContext, _ := d.GetChange("context")
	oldWriteOnly, _ := d.GetChange("context_wo")
	applied := map[string]interface{}{}
	for name := range oldContext.(map[string]interface{}) {
		applied[strings.ToUpper(name)] = nil
	}
	for name := range contextWriteOnlyNames(oldWriteOnly) {
		applied[name] = nil
	}
	for name := range context {
		applied[strings.ToUpper(name)] = nil
	}
	// Keep the vector even when it becomes empty. The VM API treats a missing
	// CONTEXT vector as no context update.
	return updateVMTemplateVec(tpl, vmk.ContextVec, applied, context)
}
