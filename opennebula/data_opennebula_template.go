package opennebula

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	templateSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/template"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaTemplateRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Template",
			},
			"cpu": func() *schema.Schema {
				s := cpuSchema()

				s.ValidateFunc = func(v interface{}, k string) (ws []string, errs []error) {
					value := v.(float64)

					if value == 0 {
						errs = append(errs, errors.New("cpu should be strictly greater than 0"))
					}

					return
				}
				return s
			}(),
			"vcpu": func() *schema.Schema {
				s := vcpuSchema()

				s.ValidateFunc = func(v interface{}, k string) (ws []string, errs []error) {
					value := v.(int)

					if value == 0 {
						errs = append(errs, errors.New("vcpu should be strictly greater than 0"))
					}

					return
				}
				return s
			}(),
			"memory": func() *schema.Schema {
				s := memorySchema()

				s.ValidateFunc = func(v interface{}, k string) (ws []string, errs []error) {
					value := v.(int)

					if value == 0 {
						errs = append(errs, errors.New("memory should be strictly greater than 0"))
					}

					return
				}
				return s
			}(),
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

	config := meta.(*Configuration)
	controller := config.Controller

	templates, err := controller.Templates().Info(-2)
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

		// Name check
		if nameOk && template.Name != name {
			continue
		}

		// CPU Check
		if cpuOk {
			tplCPU, err := template.Template.GetCPU()
			if err != nil {
				continue
			}

			if tplCPU != cpu.(float64) {
				continue
			}
		}

		// vCPU Check
		if vcpuOk {
			tplVCPU, err := template.Template.GetVCPU()
			if err != nil {
				continue
			}

			if tplVCPU != vcpu.(int) {
				continue
			}
		}

		// Memory Check
		if memoryOk {
			tplMemory, err := template.Template.GetMemory()
			if err != nil {
				continue
			}

			if memoryOk && tplMemory != memory.(int) {
				continue
			}
		}

		// Tags Check
		if tagsOk && !matchTags(template.Template.Template, tags) {
			continue
		}

		match = append(match, &templates.Templates[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no template match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several templates match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	template, err := templateFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "templates filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(template.Template.Template)

	d.SetId(strconv.FormatInt(int64(template.ID), 10))
	d.Set("name", template.Name)

	cpu, err := template.Template.GetCPU()
	if err == nil {
		d.Set("cpu", cpu)
	}

	vcpu, err := template.Template.GetVCPU()
	if err == nil {
		d.Set("vcpu", vcpu)
	}

	memory, err := template.Template.GetMemory()
	if err == nil {
		d.Set("memory", memory)
	}

	err = flattenTemplateDisks(d, &template.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "failed to flatten disks",
			Detail:   fmt.Sprintf("Template (ID: %d): %s", template.ID, err),
		})
		return diags
	}

	err = flattenTemplateNICs(d, &template.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "failed to flatten NICs",
			Detail:   fmt.Sprintf("Template (ID: %d): %s", template.ID, err),
		})
		return diags
	}

	err = flattenTemplateVMGroup(d, &template.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "failed to flatten VM groups",
			Detail:   fmt.Sprintf("Template (ID: %d): %s", template.ID, err),
		})
		return diags
	}

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("Template (ID: %d): %s", template.ID, err),
			})
			return diags
		}
	}

	return nil
}
