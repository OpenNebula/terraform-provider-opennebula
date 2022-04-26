package opennebula

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

func resourceOpennebulaGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaGroupCreate,
		Read:   resourceOpennebulaGroupRead,
		Update: resourceOpennebulaGroupUpdate,
		Delete: resourceOpennebulaGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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
				Default:     true,
				Deprecated:  "use Terraform lifcycle Meta-Argument instead.",
				Description: "Flag to delete group on destruction, by default it is set to true",
			},
			"users": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Deprecated: "use user resource for group membership instead.",
			},
			"admins": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of Admin user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Deprecated: "use opennebula_group_admins resource instead.",
			},
			"quotas": quotasSchema(),
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
			uc := controller.User(userlist[i].(int))
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
		quotasStr, err := generateQuotas(d)
		if err != nil {
			return err
		}
		err = gc.Quota(quotasStr)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaGroupRead(d, meta)
}

func resourceOpennebulaGroupRead(d *schema.ResourceData, meta interface{}) error {
	gc, err := getGroupController(d, meta)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing group %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
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
	d.Set("delete_on_destruction", d.Get("delete_on_destruction"))

	// read only configured users in current group resource
	appliedUserIDs := make([]int, 0)
	userIDsCfg := d.Get("users").([]interface{})
	for _, idCfgIf := range userIDsCfg {
		for _, id := range group.Users.ID {
			if id != idCfgIf.(int) {
				continue
			}
			appliedUserIDs = append(appliedUserIDs, id)
			break
		}
	}

	if len(appliedUserIDs) > 0 {
		err = d.Set("users", appliedUserIDs)
		if err != nil {
			return err
		}
	}

	err = d.Set("admins", group.Admins.ID)
	if err != nil {
		return err
	}
	if _, ok := d.GetOk("quotas"); ok {
		err = flattenQuotasMapFromStructs(d, &group.QuotasList)
		if err != nil {
			return err
		}
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
			quotasStr, err := generateQuotas(d)
			if err != nil {
				return err
			}
			err = gc.Quota(quotasStr)
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
