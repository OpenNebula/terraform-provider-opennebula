package opennebula

import (
	"context"
	"fmt"
	"strconv"

	hookSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/hook"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaHook() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaHookRead,
		Schema:      hookSchema(true),
	}
}

func hookFilter(d *schema.ResourceData, meta interface{}) (*hookSc.Hook, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	hooks, err := controller.Hooks().Info()
	if err != nil {
		return nil, err
	}

	id := d.Get("id")
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*hookSc.Hook, 0, 1)
	for i, hook := range hooks.Hooks {
		if id != -1 && hook.ID != id {
			continue
		}

		if nameOk && hook.Name != name {
			continue
		}

		if tagsOk && !matchTags(hook.Template.Template, tags) {
			continue
		}

		match = append(match, &hooks.Hooks[i])
	}

	if len(match) == 0 {
		return nil, fmt.Errorf("no hook matches the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several hooks match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaHookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	hook, err := hookFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "hooks filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(int64(hook.ID), 10))

	setHookResourceData(d, meta, hook)

	tags := hookTemplateTags(hook.Template.Template)
	if err := d.Set("tags", tags); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "setting attribute failed",
			Detail:   fmt.Sprintf("Hook (ID: %d): %s", hook.ID, err),
		})
		return diags
	}

	return nil
}
