package opennebula

import (
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	vmSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var sortOnVMsValues = []string{"id", "name", "cpu", "vcpu", "memory"}

func dataOpennebulaVirtualMachines() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaVirtualMachinesRead,

		Schema: map[string]*schema.Schema{
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
			"tags": tagsSchema(),
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter VMs by name with a RE2 a regular expression",
			},
			"sort_on": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Attribute used to sort the VMs list, only works on integer attributes.",

				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := strings.ToLower(v.(string))

					if !contains(value, sortOnVMsValues) {
						errors = append(errors, fmt.Errorf("type %q must be one of: %s", k, strings.Join(sortOnVMsValues, ",")))
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
			"virtual_machines": {
				Type:        schema.TypeList,
				Optional:    false,
				Computed:    true,
				Description: "List of matching vms",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Optional:    false,
							Computed:    true,
							Description: "ID of the VM",
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    false,
							Computed:    true,
							Description: "Name of the VM",
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
						"vmgroup": func() *schema.Schema {
							s := vmGroupSchema()
							s.Computed = true
							s.Optional = false
							s.MaxItems = 0
							s.Description = "Virtual Machine Group to associate with during VM creation only."
							return s
						}(),
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
	}
}

func virtualMachinesFilter(d *schema.ResourceData, meta interface{}) ([]*vmSc.VM, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	vms, err := controller.VMs().Info()
	if err != nil {
		return nil, err
	}

	// filter VMs with user defined criterias
	cpu, cpuOk := d.GetOk("cpu")
	vcpu, vcpuOk := d.GetOk("vcpu")
	memory, memoryOk := d.GetOk("memory")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})
	nameRegStr := d.Get("name_regex").(string)
	var nameReg *regexp.Regexp
	if len(nameRegStr) > 0 {
		nameReg = regexp.MustCompile(nameRegStr)
	}

	matched := make([]*vmSc.VM, 0, 1)
	for i, vm := range vms.VMs {

		if nameReg != nil && !nameReg.MatchString(vm.Name) {
			continue
		}
		tplCPU, err := vm.Template.GetCPU()
		if err != nil {
			continue
		}
		if cpuOk && tplCPU != cpu.(float64) {
			continue
		}

		tplVCPU, err := vm.Template.GetVCPU()
		if err != nil {
			continue
		}
		if vcpuOk && tplVCPU != vcpu.(int) {
			continue
		}

		tplMemory, err := vm.Template.GetMemory()
		if err != nil {
			continue
		}
		if memoryOk && tplMemory != memory.(int) {
			continue
		}

		if tagsOk && !matchTags(vm.UserTemplate.Template, tags) {
			continue
		}

		matched = append(matched, &vms.VMs[i])
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("no VMs match the constraints")
	}

	return matched, nil
}

func datasourceOpennebulaVirtualMachinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vms, err := virtualMachinesFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "VMs filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	vmsMaps := make([]map[string]interface{}, 0, len(vms))
	for _, vm := range vms {

		cpu, _ := vm.Template.GetCPU()
		vcpu, _ := vm.Template.GetVCPU()
		memory, _ := vm.Template.GetMemory()

		// builds disks list
		disks := vm.Template.GetDisks()
		diskList := make([]interface{}, 0, len(disks))

		for _, disk := range disks {
			diskList = append(diskList, flattenDisk(disk))
		}

		// builds nics list
		nics := vm.Template.GetNICs()
		nicList := make([]interface{}, 0, len(nics))

		for _, nic := range nics {
			nicList = append(nicList, flattenNIC(nic))
		}

		// builds VM Groups list
		dynTemplate := vm.Template.Template
		vmgMap := make([]map[string]interface{}, 0, 1)
		vmgIdStr, _ := dynTemplate.GetStrFromVec("VMGROUP", "VMGROUP_ID")
		vmgid, _ := strconv.ParseInt(vmgIdStr, 10, 32)
		vmgRole, _ := dynTemplate.GetStrFromVec("VMGROUP", "ROLE")

		vmgMap = append(vmgMap, map[string]interface{}{
			"vmgroup_id": vmgid,
			"role":       vmgRole,
		})

		// tags
		tplPairs := pairsToMap(vm.UserTemplate.Template)

		vmMap := map[string]interface{}{
			"name":    vm.Name,
			"id":      vm.ID,
			"cpu":     cpu,
			"vcpu":    vcpu,
			"memory":  memory,
			"disk":    diskList,
			"nic":     nicList,
			"vmgroup": vmgMap,
		}

		if len(tplPairs) > 0 {
			vmMap["tags"] = tplPairs
		}

		vmsMaps = append(vmsMaps, vmMap)
	}

	var sortOnAttr, ordering string
	nameRegStr := d.Get("name_regex").(string)
	sortOnAttr = d.Get("sort_on").(string)

	if len(sortOnAttr) > 0 && len(vmsMaps) > 1 {

		ordering = d.Get("order").(string)
		var orderingFn func(int, int) bool
		switch ordering {
		case "ASC":
			switch vmsMaps[0][sortOnAttr].(type) {
			case int:
				orderingFn = func(i, j int) bool {
					return vmsMaps[i][sortOnAttr].(int) > vmsMaps[j][sortOnAttr].(int)
				}
			case string:
				orderingFn = func(i, j int) bool {
					return vmsMaps[i][sortOnAttr].(string) > vmsMaps[j][sortOnAttr].(string)
				}
			case float64:
				orderingFn = func(i, j int) bool {
					return vmsMaps[i][sortOnAttr].(float64) > vmsMaps[j][sortOnAttr].(float64)
				}
			default:
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "can't sort VMs",
					Detail:   fmt.Sprintf("%s attribute type comparison is not handled", sortOnAttr),
				})
				return diags
			}
		case "DESC":
			switch vmsMaps[0][sortOnAttr].(type) {
			case int:
				orderingFn = func(i, j int) bool {
					return vmsMaps[i][sortOnAttr].(int) < vmsMaps[j][sortOnAttr].(int)
				}
			case string:
				orderingFn = func(i, j int) bool {
					return vmsMaps[i][sortOnAttr].(string) < vmsMaps[j][sortOnAttr].(string)
				}
			case float64:
				orderingFn = func(i, j int) bool {
					return vmsMaps[i][sortOnAttr].(float64) < vmsMaps[j][sortOnAttr].(float64)
				}
			default:
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "can't sort VMs",
					Detail:   fmt.Sprintf("%s attribute type comparison is not handled", sortOnAttr),
				})
				return diags
			}
		}

		// will crash if sortOnAttr is the name of an attributes with another type than integer
		sort.Slice(vmsMaps, func(i, j int) bool {
			return orderingFn(i, j)
		})
	}

	d.SetId(fmt.Sprintf("%x", sha512.Sum512([]byte(ordering+sortOnAttr+nameRegStr))))

	err = d.Set("virtual_machines", vmsMaps)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "failed to set virtual_machines",
			Detail:   err.Error(),
		})
		return diags

	}

	return nil
}
