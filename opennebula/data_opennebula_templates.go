package opennebula

import (
	"context"
	"crypto/sha512"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	templateSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/template"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var orderTypes = []string{"ASC", "DESC"}
var sortOnTemplatesValues = []string{"id", "name", "cpu", "vcpu", "memory", "register_date"}

func dataOpennebulaTemplates() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaTemplatesRead,

		Schema: mergeSchemas(
			commonDatasourceTemplateSchema(),
			map[string]*schema.Schema{
				"name_regex": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Filter templates by name with a RE2 a regular expression",
				},
				"sort_on": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Attribute used to sort the templates list, only works on integer attributes.",
					ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
						value := strings.ToUpper(v.(string))

						if !contains(value, sortOnTemplatesValues) {
							errors = append(errors, fmt.Errorf("type %q must be one of: %s", k, strings.Join(sortOnTemplatesValues, ",")))
						}

						return
					},
				},
				"order": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Ordering of the sort: ASC or DESC",
					ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
						value := strings.ToUpper(v.(string))

						if !contains(value, orderTypes) {
							errors = append(errors, fmt.Errorf("type %q must be one of: %s", k, strings.Join(orderTypes, ",")))
						}

						return
					},
				},
				"templates": {
					Type:        schema.TypeList,
					Optional:    false,
					Computed:    true,
					Description: "List of matching templates",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"id": {
								Type:        schema.TypeInt,
								Optional:    false,
								Computed:    true,
								Description: "ID of the Template",
							},
							"name": {
								Type:        schema.TypeString,
								Optional:    false,
								Computed:    true,
								Description: "Name of the Template",
							},
							"cpu": func() *schema.Schema {
								s := cpuSchema()
								s.Optional = false
								s.Computed = true
								return s
							}(),
							"vcpu": func() *schema.Schema {
								s := vcpuSchema()
								s.Optional = false
								s.Computed = true
								return s
							}(),
							"memory": func() *schema.Schema {
								s := memorySchema()
								s.Optional = false
								s.Computed = true
								return s
							}(),
							"disk": func() *schema.Schema {
								s := diskSchema()
								s.Computed = true
								s.Optional = false
								return s
							}(),
							"nic": func() *schema.Schema {
								s := nicSchema()
								s.Computed = true
								s.Optional = false
								return s
							}(),
							"nic_alias": func() *schema.Schema {
								s := nicAliasSchema()
								s.Computed = true
								s.Optional = false
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
							"register_date": {
								Type:        schema.TypeInt,
								Optional:    false,
								Computed:    true,
								Description: "Creation date of the template.",
							},
							"tags": func() *schema.Schema {
								s := tagsSchema()
								s.Computed = true
								s.Optional = false
								return s
							}(),
						},
					},
				},
			},
		),
	}
}

func templatesFilter(d *schema.ResourceData, meta interface{}) ([]*templateSc.Template, error) {
	newMatch := make([]*templateSc.Template, 0)

	matched, err := commonTemplatesFilter(d, meta)
	if err != nil {
		return nil, err
	}

	nameRegStr := d.Get("name_regex").(string)
	if len(nameRegStr) > 0 {
		nameReg := regexp.MustCompile(nameRegStr)
		for _, tpl := range matched {
			if !nameReg.MatchString(tpl.Name) {
				continue
			}
			newMatch = append(newMatch, tpl)
		}
	}

	if len(newMatch) == 0 {
		return nil, fmt.Errorf("no templates match the constraints")
	}

	return newMatch, nil
}

func datasourceOpennebulaTemplatesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	templates, err := templatesFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "templates filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	templatesMaps := make([]map[string]interface{}, 0, len(templates))
	for _, template := range templates {

		cpu, _ := template.Template.GetCPU()
		vcpu, _ := template.Template.GetVCPU()
		memory, _ := template.Template.GetMemory()

		// builds disks list
		disks := template.Template.GetDisks()
		diskList := make([]interface{}, 0, len(disks))

		for _, disk := range disks {
			diskList = append(diskList, flattenDisk(disk))
		}

		// builds nics list
		nics := template.Template.GetNICs()
		nicList := make([]interface{}, 0, len(nics))

		for _, nic := range nics {
			nicList = append(nicList, flattenNIC(nic))
		}

		// builds nic aliases list
		nicAliases := template.Template.GetNICAliases()
		nicAliasList := make([]interface{}, 0, len(nicAliases))

		for _, nicAlias := range nicAliases {
			nicAliasList = append(nicAliasList, flattenNICAlias(nicAlias))
		}

		// builds VM Groups list
		dynTemplate := template.Template.Template
		vmgMap := make([]map[string]interface{}, 0, 1)
		vmgIdStr, _ := dynTemplate.GetStrFromVec("VMGROUP", "VMGROUP_ID")
		vmgid, _ := strconv.ParseInt(vmgIdStr, 10, 32)
		vmgRole, _ := dynTemplate.GetStrFromVec("VMGROUP", "ROLE")

		vmgMap = append(vmgMap, map[string]interface{}{
			"vmgroup_id": vmgid,
			"role":       vmgRole,
		})

		// tags
		tplPairs := pairsToMap(template.Template.Template)

		templateMap := map[string]interface{}{
			"name":          template.Name,
			"id":            template.ID,
			"cpu":           cpu,
			"vcpu":          vcpu,
			"memory":        memory,
			"disk":          diskList,
			"nic":           nicList,
			"nic_alias":     nicAliasList,
			"vmgroup":       vmgMap,
			"register_date": template.RegTime,
		}

		if len(tplPairs) > 0 {
			templateMap["tags"] = tplPairs
		}

		templatesMaps = append(templatesMaps, templateMap)
	}

	var sortOnAttr, ordering string
	nameRegStr := d.Get("name_regex").(string)
	sortOnAttr = d.Get("sort_on").(string)

	if len(sortOnAttr) > 0 && len(templatesMaps) > 1 {
		ordering = d.Get("order").(string)
		var orderingFn func(int, int) bool
		switch ordering {
		case "ASC":
			switch templatesMaps[0][sortOnAttr].(type) {
			case int:
				orderingFn = func(i, j int) bool {
					return templatesMaps[i][sortOnAttr].(int) > templatesMaps[j][sortOnAttr].(int)
				}
			case string:
				orderingFn = func(i, j int) bool {
					return templatesMaps[i][sortOnAttr].(string) > templatesMaps[j][sortOnAttr].(string)
				}
			case float64:
				orderingFn = func(i, j int) bool {
					return templatesMaps[i][sortOnAttr].(float64) > templatesMaps[j][sortOnAttr].(float64)
				}
			default:
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "can't sort templates",
					Detail:   fmt.Sprintf("%s attribute type comparison is not handled", sortOnAttr),
				})
				return diags
			}
		case "DESC":
			switch templatesMaps[0][sortOnAttr].(type) {
			case int:
				orderingFn = func(i, j int) bool {
					return templatesMaps[i][sortOnAttr].(int) < templatesMaps[j][sortOnAttr].(int)
				}
			case string:
				orderingFn = func(i, j int) bool {
					return templatesMaps[i][sortOnAttr].(string) < templatesMaps[j][sortOnAttr].(string)
				}
			case float64:
				orderingFn = func(i, j int) bool {
					return templatesMaps[i][sortOnAttr].(float64) < templatesMaps[j][sortOnAttr].(float64)
				}
			default:
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "can't sort templates",
					Detail:   fmt.Sprintf("%s attribute type comparison is not handled", sortOnAttr),
				})
				return diags
			}
		}

		// will crash if sortOnAttr is the name of an attributes with another type than integer
		sort.Slice(templatesMaps, func(i, j int) bool {
			return orderingFn(i, j)
		})
	}

	d.SetId(fmt.Sprintf("%x", sha512.Sum512([]byte(ordering+sortOnAttr+nameRegStr))))

	err = d.Set("templates", templatesMaps)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "failed to set templates",
			Detail:   err.Error(),
		})
		return diags
	}

	return nil
}
