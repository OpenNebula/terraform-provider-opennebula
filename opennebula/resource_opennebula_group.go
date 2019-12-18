package opennebula

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	errs "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/group"
	shared "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
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
				Computed:    true,
				Description: "List of user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"admins": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of Admin user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"quotas": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "Define group quota",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datastore_quotas": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
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
							Computed:    true,
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
							Computed:    true,
							Description: "VM quotas",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:        schema.TypeFloat,
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
										Type:        schema.TypeFloat,
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
	err = d.Set("users", group.Users.ID)
	if err != nil {
		return err
	}
	err = d.Set("admins", group.Admins.ID)
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
	for _, gds := range group.Datastore {
		ds := make(map[string]interface{})
		ds["id"] = gds.ID
		ds["images"] = gds.Images
		ds["size"] = gds.Size
		datastoreQuotas = append(datastoreQuotas, ds)
	}
	// Get network quotas
	for _, gn := range group.Network {
		n := make(map[string]interface{})
		n["id"] = gn.ID
		n["leases"] = gn.Leases
		networkQuotas = append(networkQuotas, n)
	}
	// Get VM quotas
	vm := make(map[string]interface{})
	if group.VM != nil {
		vm["cpu"] = group.VM.CPU
		vm["memory"] = group.VM.Memory
		vm["running_cpu"] = group.VM.RunningCPU
		vm["running_memory"] = group.VM.RunningMemory
		vm["vms"] = group.VM.VMs
		vm["running_vms"] = group.VM.RunningVMs
		vm["system_disk_size"] = group.VM.SystemDiskSize
		vmQuotas = append(vmQuotas, vm)
	}
	// Get Image quotas
	for _, gimg := range group.Image {
		img := make(map[string]interface{})
		img["id"] = gimg.ID
		img["running_vms"] = gimg.RVMs
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

	tpl := shared.Quotas{}

	quotasMap := quotas[0].(map[string]interface{})
	datastore := quotasMap["datastore_quotas"].([]interface{})
	network := quotasMap["network_quotas"].([]interface{})
	image := quotasMap["image_quotas"].([]interface{})
	vm := quotasMap["vm_quotas"].(*schema.Set).List()

	tpl.Datastore = make([]shared.DatastoreQuota, len(datastore))
	for i := 0; i < len(datastore); i++ {
		datastoreMap := datastore[i].(map[string]interface{})

		tpl.Datastore[i] = shared.DatastoreQuota{
			ID:     datastoreMap["id"].(int),
			Images: datastoreMap["images"].(int),
			Size:   datastoreMap["size"].(int),
		}
	}

	tpl.Network = make([]shared.NetworkQuota, len(network))
	for i := 0; i < len(network); i++ {
		networkMap := network[i].(map[string]interface{})
		tpl.Network[i] = shared.NetworkQuota{
			ID:     networkMap["id"].(int),
			Leases: networkMap["leases"].(int),
		}
	}

	tpl.Image = make([]shared.ImageQuota, len(image))
	for i := 0; i < len(image); i++ {
		imageMap := image[i].(map[string]interface{})
		tpl.Image[i] = shared.ImageQuota{
			ID:   imageMap["id"].(int),
			RVMs: imageMap["running_vms"].(int),
		}
	}

	if len(vm) > 0 {
		vmMap := vm[0].(map[string]interface{})

		tpl.VM = &shared.VMQuota{
			CPU:            float32(vmMap["cpu"].(float64)),
			Memory:         vmMap["memory"].(int),
			RunningCPU:     float32(vmMap["running_cpu"].(float64)),
			RunningMemory:  vmMap["running_memory"].(int),
			RunningVMs:     vmMap["running_vms"].(int),
			SystemDiskSize: int64(vmMap["system_disk_size"].(int)),
			VMs:            vmMap["vms"].(int),
		}
	}

	w := &bytes.Buffer{}

	//Encode the VN template schema to XML
	enc := xml.NewEncoder(w)
	//enc.Indent("", "  ")
	if err := enc.Encode(tpl); err != nil {
		return ""
	}

	log.Printf("[INFO] Quotas definition: %s", w.String())
	return w.String()
}
