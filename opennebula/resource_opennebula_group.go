package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"net/http"
	"strconv"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	errs "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/group"
)

func resourceOpennebulaGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaGroupCreate,
		Read:   resourceOpennebulaGroupRead,
		Update: resourceOpennebulaGroupUpdate,
		Delete: resourceOpennebulaGroupDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Group",
			},
			"template": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Group template content, in OpenNebula XML or String format",
			},
			"delete_on_destruction": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag to delete group on destruction, by default it is set to false",
			},
			"users": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"admins": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of Admin user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"quotas": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Define group quota",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datastore_quotas": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Datastore quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Datastore ID",
									},
									"images": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of Images allowed (default: default quota)",
										Default:     -1,
									},
									"size": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum size in MB allowed on the datastore (default: default quota)",
										Default:     -1,
									},
								},
							},
						},
						"network_quotas": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Network quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Network ID",
									},
									"leases": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of Leases allowed for this network (default: default quota)",
										Default:     -1,
									},
								},
							},
						},
						"image_quotas": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Image quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Image ID",
									},
									"running_vms": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of Running VMs allowed for this image (default: default quota)",
										Default:     -1,
									},
								},
							},
						},
						"vm_quotas": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "VM quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of CPU allowed (default: default quota)",
										Default:     -1,
									},
									"memory": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum Memory (MB) allowed (default: default quota)",
										Default:     -1,
									},
									"running_cpu": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of 'running' CPUs allowed (default: default quota)",
										Default:     -1,
									},
									"running_memory": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "'Running' Memory (MB) allowed (default: default quota)",
										Default:     -1,
									},
									"running_vms": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of Running VMs allowed (default: default quota)",
										Default:     -1,
									},
									"system_disk_size": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum System Disk size (MB) allowed (default: default quota)",
										Default:     -1,
									},
									"vms": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of VMs allowed (default: default quota)",
										Default:     -1,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func getUserController(uid int, meta interface{}) (*goca.UserController, error) {
	controller := meta.(*goca.Controller)

	uc := controller.User(uid)
	if uc == nil {
		return nil, fmt.Errorf("No user with id: %d", uid)
	}
	return uc, nil
}

func getGroupController(d *schema.ResourceData, meta interface{}) (*goca.GroupController, error) {
	controller := meta.(*goca.Controller)
	var gc *goca.GroupController

	// Try to find the Group by ID, if specified
	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		gc = controller.Group(int(gid))
	}

	// Otherwise, try to find the Group by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.Groups().ByName(d.Get("name").(string))
		if err != nil {
			return nil, err
		}
		gc = controller.Group(gid)
	}

	return gc, nil
}

func resourceOpennebulaGroupCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	groupID, err := controller.Groups().Create(d.Get("name").(string))
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%v", groupID))

	gc := controller.Group(groupID)

	// add template description
	if d.Get("template") != "" {
		// Erase previous template
		err = gc.Update(d.Get("template").(string), 0)
		if err != nil {
			return err
		}
	}

	// add users if list provided
	if userids, ok := d.GetOk("users"); ok {
		userlist := userids.([]interface{})
		for i := 0; i < len(userlist); i++ {
			uc, err := getUserController(userlist[i].(int), meta)
			if err != nil {
				return err
			}
			err = uc.AddGroup(groupID)
			if err != nil {
				return err
			}
		}
	}

	// add admins if list provided
	if adminids, ok := d.GetOk("admins"); ok {
		adminlist := adminids.([]interface{})
		for i := 0; i < len(adminlist); i++ {
			err = gc.AddAdmin(adminlist[i].(int))
			if err != nil {
				return err
			}
		}
	}

	if _, ok := d.GetOk("quotas"); ok {
		err = gc.Quota(generateGroupQuotas(d))
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaGroupRead(d, meta)
}

func resourceOpennebulaGroupRead(d *schema.ResourceData, meta interface{}) error {
	gc, err := getGroupController(d, meta)
	if err != nil {
		switch err.(type) {
		case *errs.ClientError:
			clientErr, _ := err.(*errs.ClientError)
			if clientErr.Code == errs.ClientRespHTTP {
				response := clientErr.GetHTTPResponse()
				if response.StatusCode == http.StatusNotFound {
					log.Printf("[WARN] Removing group %s from state because it no longer exists in", d.Get("name"))
					d.SetId("")
					return nil
				}
			}
			return err
		default:
			return err
		}
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	group, err := gc.Info(false)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatUint(uint64(group.ID), 10))
	d.Set("name", group.Name)
	d.Set("template", group.Template)
	d.Set("delete_on_destruction", d.Get("delete_on_destruction"))
	err = d.Set("users", group.UsersID)
	if err != nil {
		return err
	}
	err = d.Set("admins", group.AdminsID)
	if err != nil {
		return err
	}
	err = flattenQuotasMapFromStructs(d, group)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	gc, err := getGroupController(d, meta)
	if err != nil {
		return err
	}

	if d.HasChange("template") {
		// Erase previous template
		err = gc.Update(d.Get("template").(string), 0)
		if err != nil {
			return err
		}
	}

	if d.HasChange("quotas") {
		if _, ok := d.GetOk("quotas"); ok {
			err = gc.Quota(generateGroupQuotas(d))
			if err != nil {
				return err
			}
		}
	}

	return resourceOpennebulaGroupRead(d, meta)
}

func resourceOpennebulaGroupDelete(d *schema.ResourceData, meta interface{}) error {
	gc, err := getGroupController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("delete_on_destruction") == true {
		err = gc.Delete()
		if err != nil {
			return err
		}
	}

	return nil
}

func flattenQuotasMapFromStructs(d *schema.ResourceData, group *group.Group) error {
	var datastoreQuotas []map[string]interface{}
	var imageQuotas []map[string]interface{}
	var vmQuotas []map[string]interface{}
	var networkQuotas []map[string]interface{}

	// Get datastore quotas
	for _, gds := range group.DatastoreQuotas {
		ds := make(map[string]interface{})
		ds["id"] = gds.ID
		images, err := strconv.ParseInt(gds.Images, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert datastore_quota.images value: %v, error: %s", gds.Images, err)
		}
		ds["images"] = images
		size, err := strconv.ParseInt(gds.Images, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert datastore_quota.size value: %v, error: %s", gds.Size, err)
		}
		ds["size"] = size
		datastoreQuotas = append(datastoreQuotas, ds)
	}
	// Get network quotas
	for _, gn := range group.NetworkQuotas {
		n := make(map[string]interface{})
		n["id"] = gn.ID
		leases, err := strconv.ParseInt(gn.Leases, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert network_quota.leases value: %v, error: %s", gn.Leases, err)
		}
		n["leases"] = leases
		networkQuotas = append(networkQuotas, n)
	}
	// Get VM quotas
	for _, gvm := range group.VMQuotas {
		vm := make(map[string]interface{})
		cpu, err := strconv.ParseInt(gvm.CPU, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert vm_quota.cpu value: %v, error: %s", gvm.CPU, err)
		}
		vm["cpu"] = cpu
		memory, err := strconv.ParseInt(gvm.Memory, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert vm_quota.memory value: %v, error: %s", gvm.Memory, err)
		}
		vm["memory"] = memory
		runningCpu, err := strconv.ParseInt(gvm.RunningCPU, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert vm_quota.running_cpu value: %v, error: %s", gvm.RunningCPU, err)
		}
		vm["running_cpu"] = runningCpu
		runningMemory, err := strconv.ParseInt(gvm.RunningMemory, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert vm_quota.running_memory value: %v, error: %s", gvm.RunningMemory, err)
		}
		vm["running_memory"] = runningMemory
		vms, err := strconv.ParseInt(gvm.VMs, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert vm_quota.vms value: %v, error: %s", gvm.VMs, err)
		}
		vm["vms"] = vms
		runningVms, err := strconv.ParseInt(gvm.RunningVMs, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert vm_quota.running_vms value: %v, error: %s", gvm.RunningVMs, err)
		}
		vm["running_vms"] = runningVms
		systemDiskSize, err := strconv.ParseInt(gvm.SystemDiskSize, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert vm_quota.system_disk_size value: %v, error: %s", gvm.SystemDiskSize, err)
		}
		vm["system_disk_size"] = systemDiskSize
		vmQuotas = append(vmQuotas, vm)
	}
	// Get Image quotas
	for _, gimg := range group.ImageQuotas {
		img := make(map[string]interface{})
		img["id"] = gimg.ID
		runningVms, err := strconv.ParseInt(gimg.RVMs, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert image_quota.running_vms value: %v, error: %s", gimg.RVMs, err)
		}
		img["running_vms"] = runningVms
		imageQuotas = append(imageQuotas, img)
	}

	return d.Set("quotas", []interface{}{
		map[string]interface{}{
			"datastore_quotas": datastoreQuotas,
			"image_quotas":     imageQuotas,
			"vm_quotas":        vmQuotas,
			"network_quotas":   networkQuotas,
		},
	})
}

func generateGroupQuotas(d *schema.ResourceData) string {
	quotas := d.Get("quotas").(*schema.Set).List()

	quotasMap := quotas[0].(map[string]interface{})
	datastore := quotasMap["datastore_quotas"].([]interface{})
	network := quotasMap["network_quotas"].([]interface{})
	image := quotasMap["image_quotas"].([]interface{})
	vm := quotasMap["vm_quotas"].(*schema.Set).List()
	quotastr := ""

	for i := 0; i < len(datastore); i++ {
		datastoreMap := datastore[i].(map[string]interface{})
		quotastr = fmt.Sprintf("%s\n%s", quotastr, fmt.Sprintf("DATASTORE = [\n  ID = %d,\n  IMAGE = %d,\n  SIZE = %d\n]",
			datastoreMap["id"].(int),
			datastoreMap["images"].(int),
			datastoreMap["size"].(int)))
	}

	for i := 0; i < len(network); i++ {
		networkMap := network[i].(map[string]interface{})
		quotastr = fmt.Sprintf("%s\n%s", quotastr, fmt.Sprintf("NETWORK = [\n  ID = %d,\n  LEASES = %d\n]",
			networkMap["id"].(int),
			networkMap["leases"].(int)))
	}

	for i := 0; i < len(image); i++ {
		imageMap := image[i].(map[string]interface{})
		quotastr = fmt.Sprintf("%s\n%s", quotastr, fmt.Sprintf("IMAGE = [\n  ID = %d,\n  RVMS = %d\n]",
			imageMap["id"].(int),
			imageMap["running_vms"].(int)))
	}

	if len(vm) > 0 {
		vmMap := vm[0].(map[string]interface{})
		quotastr = fmt.Sprintf("%s\n%s", quotastr, fmt.Sprintf("VM = [\n  CPU = %d,\n  MEMORY = %d,\n  RUNNING_CPU = %d,\n  RUNNING_MEMORY = %d,\n  RUNNING_VMS = %d,\n  SYSTEM_DISK_SIZE = %d,\n  VMS = %d\n]",
			vmMap["cpu"].(int),
			vmMap["memory"].(int),
			vmMap["running_cpu"].(int),
			vmMap["running_memory"].(int),
			vmMap["running_vms"].(int),
			vmMap["system_disk_size"].(int),
			vmMap["vms"].(int)))
	}

	return quotastr
}
