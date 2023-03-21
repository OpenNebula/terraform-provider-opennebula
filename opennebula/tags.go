package opennebula

import (
	"context"
	"fmt"
	"strings"

	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Optional:    true,
		Description: "Add custom tags to the resource",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func tagsSchemaComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Computed: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "Result of the applied default_tags and resource tags",
	}
}

func defaultTagsSchemaComputed() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "Default tags defined in the provider configuration",
	}
}

func SetTagsDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {

	newtagsAll := make(map[string]interface{}, len(meta.(*Configuration).defaultTags))

	// copy default tags map
	for k, v := range meta.(*Configuration).defaultTags {
		newtagsAll[k] = v
	}

	resourceTags := diff.Get("tags").(map[string]interface{})
	for k, v := range resourceTags {
		newtagsAll[k] = v
	}

	if err := diff.SetNew("tags_all", newtagsAll); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func SetVMTagsDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {

	newtagsAll := make(map[string]interface{}, len(meta.(*Configuration).defaultTags))

	// template_tags store uppercase keys
	templateTags := diff.Get("template_tags").(map[string]interface{})

	// copy default tags map or template key in case of override
	for k, v := range meta.(*Configuration).defaultTags {
		overrideValue, ok := templateTags[strings.ToUpper(k)]

		if !ok {
			newtagsAll[k] = v
			continue
		}
		newtagsAll[k] = overrideValue
	}

	resourceTags := diff.Get("tags").(map[string]interface{})
	for k, v := range resourceTags {
		newtagsAll[k] = v
	}

	if err := diff.SetNew("tags_all", newtagsAll); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func flattenTemplateTags(d *schema.ResourceData, meta interface{}, template *dyn.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	config := meta.(*Configuration)

	tags := make(map[string]interface{})
	tagsAll := make(map[string]interface{})

	// Get default tags
	oldDefault := d.Get("default_tags").(map[string]interface{})
	for k, _ := range oldDefault {
		tagValue, err := template.GetStr(strings.ToUpper(k))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Failed to get default tag",
			})
		}
		tagsAll[k] = tagValue
	}
	d.Set("default_tags", config.defaultTags)

	// Get only tags described in the configuration
	if tagsInterface, ok := d.GetOk("tags"); ok {

		for k, _ := range tagsInterface.(map[string]interface{}) {
			key := strings.ToUpper(k)
			tagValue, err := template.GetStr(key)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Failed to get tag from the template",
				})
			}
			tags[k] = tagValue
			tagsAll[k] = tagValue
		}

		err := d.Set("tags", tags)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Failed to set attribute",
			})
		}
	}
	d.Set("tags_all", tagsAll)

	return nil
}
