package opennebula

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type DefaultTags struct {
	Elements map[string]string
}

func (t *DefaultTags) FromTerraform5Value(val tftypes.Value) error {

	v := map[string]tftypes.Value{}
	err := val.As(&v)
	if err != nil {
		return err
	}

	tmpTags := make(map[string]tftypes.Value)

	err = v["tags"].As(&tmpTags)
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

//type defaultTagsModifier struct {
//	tags map[string]attr.Value
//}
//
//func (d defaultTagsModifier) Description(ctx context.Context) string {
//	return fmt.Sprintf("Applies default tags then override with resource tags to produce the new plan")
//}
//
//func (d defaultTagsModifier) MarkdownDescription(ctx context.Context) string {
//	return fmt.Sprintf("Applies default tags then override with resource tags to produce the new plan")
//
//}
//
//// PlanModifyString runs the logic of the plan modifier.
//// Access to the configuration, plan, and state is available in `req`, while
//// `resp` contains fields for updating the planned value, triggering resource
//// replacement, and returning diagnostics.
//func (d defaultTagsModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
//	//req.Plan
//
//	resp.PlanValue, resp.Diagnostics = types.MapValue(types.MapType{}, d.tags)
//}
//
//func defaultTagsModifierInit() *defaultTagsModifier {
//	return &defaultTagsModifier{}
//}
//
