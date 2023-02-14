package common

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TemplateSectionBlock() schema.Block {
	return schema.SetNestedBlock{
		Description: "Add default tags to the resources",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the section",
					//MarkdownDescription: "",
				},
				"elements": schema.MapAttribute{
					Optional:    true,
					Description: "Tags of the section",
					ElementType: types.StringType,
					//MarkdownDescription: "",
				},
			},
		},
	}
}

type TemplateSection struct {
	Name     string
	Elements map[string]string
}

func (t *TemplateSection) FromTerraform5Value(val tftypes.Value) error {

	// Get tags representation as golang types
	v := map[string]tftypes.Value{}
	err := val.As(&v)
	if err != nil {
		return err
	}

	// get name
	err = v["name"].As(&t.Name)
	if err != nil {
		return err
	}

	// get section elements
	tmpTags := make(map[string]tftypes.Value)

	err = v["elements"].As(&tmpTags)
	if err != nil {
		return err
	}

	t.Elements = make(map[string]string)

	for k, v := range tmpTags {
		if v.Type().Is(tftypes.String) {
			value := ""
			v.As(&value)
			t.Elements[k] = value
		}
	}

	return nil
}
