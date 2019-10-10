package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
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
						"datastore": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Datastore quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"datastore_id": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Datastore ID",
									},
									"images": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of Images allowed (default: unlimited)",
										Default:     -2,
									},
									"size": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum size in MB allowed on the datastore (default: unlimited)",
										Default:     -2,
									},
								},
							},
						},
						"network": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Network quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_id": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Network ID",
									},
									"leases": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of Leases allowed for this network (default: unlimited)",
										Default:     -2,
									},
								},
							},
						},
						"image": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Image quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_id": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Image ID",
									},
									"running_vms": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of Running VMs allowed for this image (default: unlimited)",
										Default:     -2,
									},
								},
							},
						},
						"vm": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "VM quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of CPU allowed (default: unlimited)",
										Default:     -2,
									},
									"memory": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum Memory (MB) allowed (default: unlimited)",
										Default:     -2,
									},
									"running_cpu": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of 'running' CPUs allowed (default: unlimited)",
										Default:     -2,
									},
									"running_memory": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "'Running' Memory (MB) allowed (default: unlimited)",
										Default:     -2,
									},
									"running_vms": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of Running VMs allowed (default: unlimited)",
										Default:     -2,
									},
									"system_disk_size": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum System Disk size (MB) allowed (default: unlimited)",
										Default:     -2,
									},
									"vms": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Maximum number of VMs allowed (default: unlimited)",
										Default:     -2,
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

func getGroupController(d *schema.ResourceData, meta interface{}) (*goca.GroupController, error) {
	controller := meta.(*goca.Controller)
	var gc *goca.GroupController

	// Try to find the Group by ID, if specified
	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Group Id (%s) is not an integer", d.Id())
		}
		gc = controller.Group(int(gid))
	}

	// Otherwise, try to find the Group by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.Groups().ByName(d.Get("name").(string))
		if err != nil {
			d.SetId("")
			return nil, fmt.Errorf("Could not find Group with name %s", d.Get("name").(string))
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
		return err
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

func generateGroupQuotas(d *schema.ResourceData) string {
	quotas := d.Get("quotas").(*schema.Set).List()

	quotasMap := quotas[0].(map[string]interface{})
	datastore := quotasMap["datastore"].([]interface{})
	network := quotasMap["network"].([]interface{})
	image := quotasMap["image"].([]interface{})
	vm := quotasMap["vm"].(*schema.Set).List()
	quotastr := ""

	for i := 0; i < len(datastore); i++ {
		datastoreMap := datastore[i].(map[string]interface{})
		quotastr = fmt.Sprintf("%s\n%s", quotastr, fmt.Sprintf("DATASTORE = [\n  ID = %d,\n  IMAGE = %d,\n  SIZE = %d\n]",
			datastoreMap["datastore_id"].(int),
			datastoreMap["images"].(int),
			datastoreMap["size"].(int)))
	}

	for i := 0; i < len(network); i++ {
		networkMap := network[i].(map[string]interface{})
		quotastr = fmt.Sprintf("%s\n%s", quotastr, fmt.Sprintf("NETWORK = [\n  ID = %d,\n  LEASES = %d\n]",
			networkMap["network_id"].(int),
			networkMap["leases"].(int)))
	}

	for i := 0; i < len(image); i++ {
		imageMap := image[i].(map[string]interface{})
		quotastr = fmt.Sprintf("%s\n%s", quotastr, fmt.Sprintf("IMAGE = [\n  ID = %d,\n  RVMS = %d\n]",
			imageMap["image_id"].(int),
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
