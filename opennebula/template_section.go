package opennebula

import (
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func templateSectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "Add custom section to the resource",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"elements": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func addTemplateVectors(vectorsInterface []interface{}, tpl *dyn.Template) {

	for _, vectorIf := range vectorsInterface {
		vector := vectorIf.(map[string]interface{})
		vecName := strings.ToUpper(vector["name"].(string))
		vecElements := vector["elements"].(map[string]interface{})

		vec := tpl.AddVector(strings.ToUpper(vecName))
		for k, v := range vecElements {
			vec.AddPair(k, v)
		}
	}
}

func flattenTemplateSection(d *schema.ResourceData, meta interface{}, tpl *dynamic.Template) error {

	if vectorsInterface, ok := d.GetOk("template_section"); ok {

		templateSection := make([]interface{}, 0)

		for _, vectorIf := range vectorsInterface.(*schema.Set).List() {
			vector := vectorIf.(map[string]interface{})
			vecName := vector["name"].(string)
			vecElements := vector["elements"].(map[string]interface{})

			// Suppose vector key unicity
			vectorTpl, err := tpl.GetVector(strings.ToUpper(vecName))
			if err != nil {
				continue
			}

			elements := make(map[string]interface{})
			for _, pair := range vectorTpl.Pairs {
				for k, _ := range vecElements {
					if strings.ToUpper(k) != pair.Key() {
						continue
					}
					elements[k] = pair.Value
					break
				}
			}

			templateSection = append(templateSection, map[string]interface{}{
				"name":     vecName,
				"elements": elements,
			})

		}

		err := d.Set("template_section", templateSection)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateTemplateSection(d *schema.ResourceData, newTpl *dyn.Template) {

	oldVectorsIf, newVectorsIf := d.GetChange("template_section")
	oldVectors := oldVectorsIf.(*schema.Set).List()
	newVectors := newVectorsIf.(*schema.Set).List()

	// Here we suppose vector key unicity
	// delete vectors
	for _, oldVectorIf := range oldVectors {
		oldVector := oldVectorIf.(map[string]interface{})
		oldVectorName := oldVector["name"].(string)

		if len(newVectors) == 0 {
			newTpl.Del(strings.ToUpper(oldVectorName))
		}

		// if a new vector has the same name, keep it
		for _, newVectorIf := range newVectors {
			newVector := newVectorIf.(map[string]interface{})

			if oldVectorName == newVector["name"].(string) {
				continue
			}
			newTpl.Del(strings.ToUpper(oldVectorName))
		}

	}

	// add/update vectors
	for _, newVectorIf := range newVectors {
		newVector := newVectorIf.(map[string]interface{})
		newVectorName := strings.ToUpper(newVector["name"].(string))

		elements := newVector["elements"].(map[string]interface{})

		// it seems there is a problem, sometimes an empty map is present in newVectors list
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/588
		if len(newVectorName) == 0 && len(elements) == 0 {
			continue
		}

		newTpl.Del(strings.ToUpper(newVectorName))
		newVec := newTpl.AddVector(newVectorName)
		for k, v := range elements {
			newVec.AddPair(k, v)
		}
	}
}
