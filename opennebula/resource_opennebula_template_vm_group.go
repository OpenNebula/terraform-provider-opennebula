package opennebula

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	errs "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vmgroup"
)

func resourceOpennebulaVMGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaVMGroupCreate,
		Read:   resourceOpennebulaVMGroupRead,
		Exists: resourceOpennebulaVMGroupExists,
		Update: resourceOpennebulaVMGroupUpdate,
		Delete: resourceOpennebulaVMGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the template",
			},
			"role": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "Roles of the VM Group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"host_affined": {
							Type:        schema.TypeList,
							Computed:    true,
							Optional:    true,
							Description: "Host(s) where the VM should run",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"host_anti_affined": {
							Type:        schema.TypeList,
							Computed:    true,
							Optional:    true,
							Description: "Host(s) where the VM should not run",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"policy": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								validpolicies := []string{"NONE", "AFFINED", "ANTI_AFFINED"}
								value := v.(string)

								if inArray(value, validpolicies) < 0 {
									errors = append(errors, fmt.Errorf("Policy value %q must be one of: %s", k, strings.Join(validpolicies, ",")))
								}

								return
							},
						},
						"vms": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
							Description: "VMs using VM Group",
						},
					},
				},
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the template vm group (in Unix format, owner-group-other, use-manage-admin)",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if len(value) != 3 {
						errors = append(errors, fmt.Errorf("%q has specify 3 permission sets: owner-group-other", k))
					}

					all := true
					for _, c := range strings.Split(value, "") {
						if c < "0" || c > "7" {
							all = false
						}
					}
					if !all {
						errors = append(errors, fmt.Errorf("Each character in %q should specify a Unix-like permission set with a number from 0 to 7", k))
					}

					return
				},
			},

			"uid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the user that will own the template vm group",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the template vm group",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the template vm group",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the template vm group",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the Template VM Group, If empty, it uses caller group",
			},
		},
	}
}

func getVMGroupController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.VMGroupController, error) {
	controller := meta.(*goca.Controller)
	var vmgc *goca.VMGroupController

	// Try to find the vm group by ID, if specified
	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		vmgc = controller.VMGroup(int(gid))
	}

	// Otherwise, try to find the template by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.VMGroups().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
		}
		vmgc = controller.VMGroup(gid)
	}

	return vmgc, nil
}

func changeVMGroupGroup(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var gid int

	vmgc, err := getVMGroupController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		gid, err = controller.Groups().ByName(d.Get("group").(string))
		if err != nil {
			return err
		}
	}

	err = vmgc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaVMGroupCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	vmg, xmlerr := generateVMGroup(d)
	if xmlerr != nil {
		return xmlerr
	}

	vmgID, err := controller.VMGroups().Create(vmg)
	if err != nil {
		log.Printf("[ERROR] VMGroup creation failed, error: %s", err)
		return err
	}

	vmgc := controller.VMGroup(vmgID)

	d.SetId(fmt.Sprintf("%v", vmgID))

	// Change Permissions only if Permissions are set
	if perms, ok := d.GetOk("permissions"); ok {
		p := permissionUnix(perms.(string))
		err = vmgc.Chmod(p.OwnerU, p.OwnerM, p.OwnerA, p.GroupU, p.GroupM, p.GroupA, p.OtherU, p.OwnerM, p.OtherA)
		if err != nil {
			log.Printf("[ERROR] template permissions change failed, error: %s", err)
			return err
		}
	}

	if d.Get("group") != "" {
		err = changeVMGroupGroup(d, meta)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaVMGroupRead(d, meta)
}

func resourceOpennebulaVMGroupRead(d *schema.ResourceData, meta interface{}) error {
	// Get requested template from all templates
	vmgc, err := getVMGroupController(d, meta, -2, -1, -1)
	if err != nil {
		switch err.(type) {
		case *errs.ClientError:
			clientErr, _ := err.(*errs.ClientError)
			if clientErr.Code == errs.ClientRespHTTP {
				response := clientErr.GetHTTPResponse()
				if response.StatusCode == http.StatusNotFound {
					log.Printf("[WARN] Removing virtual machine template %s from state because it no longer exists in", d.Get("name"))
					d.SetId("")
					return nil
				}
			}
			return err
		default:
			return err
		}
	}

	vmg, err := vmgc.Info(false)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", vmg.ID))
	d.Set("name", vmg.Name)
	d.Set("uid", vmg.UID)
	d.Set("gid", vmg.GID)
	d.Set("uname", vmg.UName)
	d.Set("gname", vmg.GName)
	d.Set("permissions", permissionsUnixString(vmg.Permissions))

	// Get Human readable vmg information
	err = flattenVMGroupRoles(d, vmg.Roles)
	if err != nil {
		return err
	}

	return nil
}

func flattenVMGroupRoles(d *schema.ResourceData, vmgRoles []vmgroup.Role) error {
	var roles []map[string]interface{}

	for _, vmgr := range vmgRoles {

		hAff := make([]int, 0)
		hAntiAff := make([]int, 0)
		vms := make([]int, 0)
		if vmgr.HostAffined != "" {
			hostAffString := strings.Split(vmgr.HostAffined, ",")
			for _, h := range hostAffString {
				hostAffInt, _ := strconv.ParseInt(h, 10, 32)
				hAff = append(hAff, int(hostAffInt))
			}
		}
		if vmgr.HostAntiAffined != "" {
			hostAntiAffString := strings.Split(vmgr.HostAntiAffined, ",")
			for _, h := range hostAntiAffString {
				hostAntiAffInt, _ := strconv.ParseInt(h, 10, 32)
				hAntiAff = append(hAff, int(hostAntiAffInt))
			}
		}
		if vmgr.VMs != "" {
			vmsString := strings.Split(vmgr.VMs, ",")
			for _, vm := range vmsString {
				vmInt, _ := strconv.ParseInt(vm, 10, 32)
				vms = append(vms, int(vmInt))
			}
		}
		roles = append(roles, map[string]interface{}{
			"id":                vmgr.ID,
			"name":              vmgr.Name,
			"host_affined":      hAff,
			"host_anti_affined": hAntiAff,
			"policy":            vmgr.Policy,
			"vms":               vms,
		})
	}

	return d.Set("role", roles)
}

func resourceOpennebulaVMGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaVMGroupRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaVMGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	//Get VMGroup
	vmgc, err := getVMGroupController(d, meta)
	if err != nil {
		return err
	}
	vmg, err := vmgc.Info(false)
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		err := vmgc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
		vmg, err = vmgc.Info(false)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated name for vmg %s\n", vmg.Name)
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			p := permissionUnix(perms.(string))
			err = vmgc.Chmod(p.OwnerU, p.OwnerM, p.OwnerA, p.GroupU, p.GroupM, p.GroupA, p.OtherU, p.OtherM, p.OtherA)
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated VMGroup %s\n", vmg.Name)
	}

	if d.HasChange("group") {
		err = changeVMGroupGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for VMGroup %s\n", vmg.Name)
	}

	if d.HasChange("role") && d.Get("role") != "" {
		var err error

		vmg, xmlerr := generateVMGroup(d)
		if xmlerr != nil {
			return xmlerr
		}

		err = vmgc.Update(vmg, 0)
		if err != nil {
			return err
		}

		log.Printf("[INFO] Successfully updated Virtual Machine Group %s\n", d.Id())
	}

	return nil
}

func resourceOpennebulaVMGroupDelete(d *schema.ResourceData, meta interface{}) error {
	vmgc, err := getVMGroupController(d, meta)
	if err != nil {
		return err
	}

	err = vmgc.Delete()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted VMGroup ID %s\n", d.Id())

	return nil
}

func generateVMGroup(d *schema.ResourceData) (string, error) {

	tpl := dynamic.NewTemplate()

	tpl.AddPair("NAME", d.Get("name").(string))

	// Add Roles to the template
	roles := d.Get("role").([]interface{})
	for _, r := range roles {

		role := r.(map[string]interface{})
		rHostAff := ArrayToString(role["host_affined"].([]interface{}), ",")
		rHostAntiAff := ArrayToString(role["host_anti_affined"].([]interface{}), ",")

		roleTpl := tpl.AddVector("ROLE")
		roleTpl.AddPair("NAME", role["name"].(string))
		roleTpl.AddPair("HOST_AFFINED", rHostAff)
		roleTpl.AddPair("HOST_ANTI_AFFINED", rHostAntiAff)
		roleTpl.AddPair("POLICY", role["policy"].(string))

	}

	tplStr := tpl.String()
	log.Printf("[INFO] VM group definition: %s", tplStr)

	return tplStr, nil
}
