package opennebula

import "github.com/hashicorp/terraform-plugin-go/tftypes"

type Tags struct {
	elements map[string]string
}

func (t *Tags) FromTerraform5Value(val tftypes.Value) error {

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

	t.elements = make(map[string]string)

	for k, v := range tmpTags {
		if v.Type().Is(tftypes.String) {
			value := ""
			v.As(&value)
			t.elements[k] = value
		}
	}

	return nil
}
