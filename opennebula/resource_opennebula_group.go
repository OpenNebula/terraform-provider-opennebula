package opennebula

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
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
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Group template content, in OpenNebula XML or String format",
				Deprecated:    "use other schema sections",
				ConflictsWith: []string{"sunstone", "tags"},
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
			"sunstone": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Allow users and group admins to access specific views",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_view": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Default Sunstone view for regular users",
						},
						"views": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "List of available views for regular users",
						},
						"group_admin_default_view": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Default Sunstone view for group admin users",
						},
						"group_admin_views": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "List of available views for the group admins",
						},
					},
				},
				ConflictsWith: []string{"template"},
			},
			"tags": func() *schema.Schema {
				s := tagsSchema()
				s.ConflictsWith = []string{"template"}
				return s
			}(),
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

	// template management

	// add template description
	if d.Get("template") != "" {
		// Erase previous template
		err = gc.Update(d.Get("template").(string), 0)
		if err != nil {
			return err
		}
	}

	tpl := dyn.NewTemplate()

	sunstone := d.Get("sunstone").(*schema.Set).List()
	if len(sunstone) > 0 {
		sunstoneVec := makeSunstoneVec(sunstone[0].(map[string]interface{}))
		tpl.Elements = append(tpl.Elements, sunstoneVec)
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
	}

	if len(tpl.Elements) > 0 {
		err = gc.Update(tpl.String(), parameters.Merge)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaGroupRead(d, meta)
}

func makeSunstoneVec(sunstoneConfig map[string]interface{}) *dyn.Vector {

	vector := dyn.Vector{
		XMLName: xml.Name{Local: "SUNSTONE"},
	}

	defaultView := sunstoneConfig["default_view"].(string)
	if len(defaultView) > 0 {
		vector.AddPair("DEFAULT_VIEW", defaultView)
	}

	views := sunstoneConfig["views"].(string)
	if len(views) > 0 {
		vector.AddPair("VIEWS", views)
	}

	groupAdminDefaultView := sunstoneConfig["group_admin_default_view"].(string)
	if len(groupAdminDefaultView) > 0 {
		vector.AddPair("GROUP_ADMIN_DEFAULT_VIEW", groupAdminDefaultView)
	}

	groupAdminViews := sunstoneConfig["group_admin_views"].(string)
	if len(groupAdminViews) > 0 {
		vector.AddPair("GROUP_ADMIN_VIEWS", groupAdminViews)
	}

	return &vector
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
	d.Set("template", group.Template.String())
	deleteOnDestruction, ok := d.Get("delete_on_destruction").(bool)
	if ok {
		d.Set("delete_on_destruction", deleteOnDestruction)
	}

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

	err = flattenGroupTemplate(d, &group.Template)
	if err != nil {
		return err
	}

	return nil
}

func flattenGroupTemplate(d *schema.ResourceData, groupTpl *dyn.Template) error {

	tags := make(map[string]interface{})
	for i, _ := range groupTpl.Elements {

		switch e := groupTpl.Elements[i].(type) {
		case *dyn.Pair:

			// Get only tags described in the configuration
			if tagsInterface, ok := d.GetOk("tags"); ok {
				var err error
				for k, _ := range tagsInterface.(map[string]interface{}) {
					tags[k], err = groupTpl.GetStr(strings.ToUpper(k))
					if err != nil {
						return err
					}
				}

			}
		case *dyn.Vector:
			switch e.Key() {
			case "SUNSTONE":
				defaultView, _ := e.GetStr("DEFAULT_VIEW")
				views, _ := e.GetStr("VIEWS")
				groupAdminDefaultView, _ := e.GetStr("GROUP_ADMIN_DEFAULT_VIEW")
				groupAdminViews, _ := e.GetStr("GROUP_ADMIN_VIEWS")

				sunstoneConfig := []map[string]interface{}{
					{
						"default_view":             defaultView,
						"views":                    views,
						"group_admin_default_view": groupAdminDefaultView,
						"group_admin_views":        groupAdminViews,
					},
				}

				err := d.Set("sunstone", sunstoneConfig)
				if err != nil {
					return err
				}
			default:
				log.Printf("[DEBUG] ignored: %s", e)
			}

		}

	}

	if len(tags) > 0 {
		err := d.Set("tags", tags)
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

	// template management

	group, err := gc.Info(false)
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

	update := false
	newTpl := group.Template

	if d.HasChange("sunstone") {
		newTpl.Del("SUNSTONE")

		sunstone := d.Get("sunstone").(*schema.Set).List()
		if len(sunstone) > 0 {
			sunstoneVec := makeSunstoneVec(sunstone[0].(map[string]interface{}))
			newTpl.Elements = append(newTpl.Elements, sunstoneVec)
		}

		update = true
	}

	if d.HasChange("tags") {

		oldTagsIf, newTagsIf := d.GetChange("tags")
		oldTags := oldTagsIf.(map[string]interface{})
		newTags := newTagsIf.(map[string]interface{})

		// delete tags
		for k, _ := range oldTags {
			_, ok := newTags[k]
			if ok {
				continue
			}
			newTpl.Del(strings.ToUpper(k))
		}

		// add/update tags
		for k, v := range newTags {
			newTpl.Del(strings.ToUpper(k))
			newTpl.AddPair(strings.ToUpper(k), v)
		}

		update = true
	}

	if update {
		err = gc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			return err
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
