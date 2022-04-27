package opennebula

import (
	"fmt"
	"strconv"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	templateSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/template"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaTemplate() *schema.Resource {
	return &schema.Resource{
		Read: datasourceOpennebulaTemplateRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Template",
			},
			"cpu":    cpuSchema(),
			"vcpu":   vcpuSchema(),
			"memory": memorySchema(),
			"context": func() *schema.Schema {
				s := contextSchema()
				s.Deprecated = "use 'tags' for selection instead"
				return s
			}(),
			"disk": func() *schema.Schema {
				s := diskSchema()
				s.Computed = true
				s.Optional = false
				return s
			}(),
			"graphics": func() *schema.Schema {
				s := graphicsSchema()
				s.Deprecated = "use 'tags' for selection instead"
				return s
			}(),
			"nic": func() *schema.Schema {
				s := nicSchema()
				s.Computed = true
				s.Optional = false
				return s
			}(),
			"os": func() *schema.Schema {
				s := osSchema()
				s.Deprecated = "use 'tags' for selection instead"
				return s
			}(),
			"vmgroup": func() *schema.Schema {
				s := vmGroupSchema()
				s.Computed = true
				s.Optional = false
				s.MaxItems = 0
				s.Description = "Virtual Machine Group to associate with during VM creation only."
				return s
			}(),
			"tags": tagsSchema(),
		},
	}
}

func templateFilter(d *schema.ResourceData, meta interface{}) (*templateSc.Template, error) {

	controller := meta.(*goca.Controller)

	templates, err := controller.Templates().Info()
	if err != nil {
		return nil, err
	}

	// filter templates with user defined criterias
	name, nameOk := d.GetOk("name")
	cpu, cpuOk := d.GetOk("cpu")
	vcpu, vcpuOk := d.GetOk("vcpu")
	memory, memoryOk := d.GetOk("memory")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*templateSc.Template, 0, 1)
	for i, template := range templates.Templates {

		if nameOk && template.Name != name {
			continue
		}

		tplCPU, err := template.Template.GetCPU()
		if err != nil {
			continue
		}

		if cpuOk && tplCPU != cpu.(float64) {
			continue
		}

		tplVCPU, err := template.Template.GetVCPU()
		if err != nil {
			continue
		}

		if vcpuOk && tplVCPU != vcpu.(int) {
			continue
		}

		tplMemory, err := template.Template.GetMemory()
		if err != nil {
			continue
		}

		if memoryOk && tplMemory != memory.(int) {
			continue
		}

		if tagsOk && !matchTags(template.Template.Template, tags) {
			continue
		}

		match = append(match, &templates.Templates[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no template match the tags")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several templates match the tags")
	}

	return match[0], nil
}

func datasourceOpennebulaTemplateRead(d *schema.ResourceData, meta interface{}) error {

	template, err := templateFilter(d, meta)
	if err != nil {
		return err
	}

	tplPairs := pairsToMap(template.Template.Template)

	d.SetId(strconv.FormatInt(int64(template.ID), 10))
	d.Set("name", template.Name)

	cpu, err := template.Template.GetCPU()
	if err != nil {
		return err
	}
	d.Set("cpu", cpu)

	vcpu, err := template.Template.GetVCPU()
	if err != nil {
		return err
	}
	d.Set("vcpu", vcpu)

	memory, err := template.Template.GetMemory()
	if err != nil {
		return err
	}
	d.Set("memory", memory)

	err = flattenTemplateDisks(d, &template.Template)
	if err != nil {
		return err
	}

	err = flattenTemplateNICs(d, &template.Template)
	if err != nil {
		return err
	}

	err = flattenTemplateVMGroup(d, &template.Template)
	if err != nil {
		return err
	}

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			return err
		}
	}

	return nil
}
