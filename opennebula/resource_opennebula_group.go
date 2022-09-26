package opennebula

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
)

func resourceOpennebulaGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaGroupCreate,
		ReadContext:   resourceOpennebulaGroupRead,
		UpdateContext: resourceOpennebulaGroupUpdate,
		DeleteContext: resourceOpennebulaGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Group",
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
			},
			"tags":         tagsSchema(),
			"default_tags": defaultTagsSchemaComputed(),
			"tags_all":     tagsSchemaComputed(),
		},
	}
}

func getGroupController(d *schema.ResourceData, meta interface{}) (*goca.GroupController, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	var gc *goca.GroupController

	// Try to find the Group by ID, if specified
	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 0)
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

func resourceOpennebulaGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID, err := controller.Groups().Create(d.Get("name").(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the group",
			Detail:   err.Error(),
		})
		return diags
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
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add users",
					Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
				})
				return diags
			}
		}
	}

	// add admins if list provided
	if adminids, ok := d.GetOk("admins"); ok {
		adminlist := adminids.([]interface{})
		for i := 0; i < len(adminlist); i++ {
			err = gc.AddAdmin(adminlist[i].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add admins",
					Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
				})
				return diags
			}
		}
	}

	if _, ok := d.GetOk("quotas"); ok {
		quotasStr, err := generateQuotas(d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate quotas description",
				Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
			})
			return diags
		}
		err = gc.Quota(quotasStr)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to apply quotas",
				Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
			})
			return diags
		}
	}

	// template management

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

	// add default tags if they aren't overriden
	if len(config.defaultTags) > 0 {
		for k, v := range config.defaultTags {
			key := strings.ToUpper(k)
			p, _ := tpl.GetPair(key)
			if p != nil {
				continue
			}
			tpl.AddPair(key, v)
		}
	}

	if len(tpl.Elements) > 0 {
		err = gc.Update(tpl.String(), parameters.Merge)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update the group content",
				Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
			})
			return diags
		}
	}

	return resourceOpennebulaGroupRead(ctx, d, meta)
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

func resourceOpennebulaGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	gc, err := getGroupController(d, meta)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing group %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the group controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	group, err := gc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed retrieve group informations",
			Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(strconv.FormatUint(uint64(group.ID), 10))
	d.Set("name", group.Name)

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
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed set field",
				Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	err = d.Set("admins", group.Admins.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set field",
			Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	if _, ok := d.GetOk("quotas"); ok {
		err = flattenQuotasMapFromStructs(d, &group.QuotasList)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to flatten quotas",
				Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	err = flattenGroupTemplate(d, meta, &group.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten template",
			Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return nil
}

func flattenGroupTemplate(d *schema.ResourceData, meta interface{}, groupTpl *dyn.Template) error {
	config := meta.(*Configuration)

	for i, _ := range groupTpl.Elements {

		switch e := groupTpl.Elements[i].(type) {

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

	tags := make(map[string]interface{})
	tagsAll := make(map[string]interface{})

	// Get default tags
	oldDefault := d.Get("default_tags").(map[string]interface{})
	for k, _ := range oldDefault {
		tagValue, err := groupTpl.GetStr(strings.ToUpper(k))
		if err != nil {
			return nil
		}
		tagsAll[k] = tagValue
	}
	d.Set("default_tags", config.defaultTags)

	// Get only tags described in the configuration
	if tagsInterface, ok := d.GetOk("tags"); ok {

		for k, _ := range tagsInterface.(map[string]interface{}) {
			tagValue, err := groupTpl.GetStr(strings.ToUpper(k))
			if err != nil {
				return err
			}
			tags[k] = tagValue
			tagsAll[k] = tagValue
		}

		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
	}
	d.Set("tags_all", tagsAll)

	return nil
}

func resourceOpennebulaGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	gc, err := getGroupController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the group controller",
			Detail:   err.Error(),
		})
		return diags
	}

	if d.HasChange("quotas") {
		if _, ok := d.GetOk("quotas"); ok {
			quotasStr, err := generateQuotas(d)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to generate quotas",
					Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

			err = gc.Quota(quotasStr)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to apply quotas",
					Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

		}
	}

	// template management

	group, err := gc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve groups informations",
			Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
		})
		return diags
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
			key := strings.ToUpper(k)
			newTpl.Del(key)
			newTpl.AddPair(key, v)
		}

		update = true
	}

	if d.HasChange("tags_all") {
		oldTagsAllIf, newTagsAllIf := d.GetChange("tags_all")
		oldTagsAll := oldTagsAllIf.(map[string]interface{})
		newTagsAll := newTagsAllIf.(map[string]interface{})

		tags := d.Get("tags").(map[string]interface{})

		// delete tags
		for k, _ := range oldTagsAll {
			_, ok := newTagsAll[k]
			if ok {
				continue
			}
			newTpl.Del(strings.ToUpper(k))
		}

		// reapply all default tags that were neither applied nor overriden via tags section
		for k, v := range newTagsAll {
			_, ok := tags[k]
			if ok {
				continue
			}

			key := strings.ToUpper(k)
			newTpl.Del(key)
			newTpl.AddPair(key, v)
		}

		update = true
	}

	if update {
		err = gc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update group content",
				Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

	}

	return resourceOpennebulaGroupRead(ctx, d, meta)
}

func resourceOpennebulaGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	gc, err := getGroupController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the group controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// Group should be empty to be removed
	groupInfos, err := gc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed retrieve group informations",
			Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	for _, userID := range groupInfos.Users.ID {

		userInfos, err := controller.User(userID).Info(false)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to get user informations",
				Detail:   fmt.Sprintf("group (ID: %d) user (ID: %d): %s", gc.ID, userID, err),
			})
			return diags
		}

		if userInfos.GID == gc.ID {
			// It's a primary group: we need to move the user to a default group, here we move to "users"
			err := controller.User(userID).Chgrp(1)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add user to users group",
					Detail:   fmt.Sprintf("group (ID: %d) user (ID: %d): %s", gc.ID, userID, err),
				})
				return diags
			}
		} else {
			// It's a secondary group, we just remove it
			err := controller.User(userID).DelGroup(gc.ID)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add user to users group",
					Detail:   fmt.Sprintf("group (ID: %d) user (ID: %d): %s", gc.ID, userID, err),
				})
				return diags
			}
		}
	}

	err = gc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("group (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return nil
}
