package opennebula

import (
	"fmt"
	"strconv"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	groupSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/group"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataOpennebulaGroup() *schema.Resource {
	return &schema.Resource{
		Read: datasourceOpennebulaGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the group",
			},
			"admins": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Admin user IDs part of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"quotas": func() *schema.Schema {
				s := quotasSchema()
				s.Deprecated = "use 'tags' for selection instead"
				return s
			}(),
			"tags": tagsSchema(),
		},
	}
}

func groupFilter(d *schema.ResourceData, meta interface{}) (*groupSc.GroupShort, error) {

	controller := meta.(*goca.Controller)

	groups, err := controller.Groups().Info()
	if err != nil {
		return nil, err
	}

	// filter groups with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*groupSc.GroupShort, 0, 1)
	for i, group := range groups.Groups {

		if nameOk && group.Name != name {
			continue
		}

		if tagsOk && !matchTags(group.Template, tags) {
			continue
		}

		match = append(match, &groups.Groups[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no group match the tags")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several groups match the tags")
	}

	return match[0], nil
}

func datasourceOpennebulaGroupRead(d *schema.ResourceData, meta interface{}) error {

	controller := meta.(*goca.Controller)

	group, err := groupFilter(d, meta)
	if err != nil {
		return err
	}

	tplPairs := pairsToMap(group.Template)

	d.SetId(strconv.FormatInt(int64(group.ID), 10))
	d.Set("name", group.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			return err
		}
	}

	// read only configured users in current group resource
	userIDsCfg, ok := d.GetOk("users")
	if ok {
		appliedUserIDs := make([]int, 0)

		for _, idCfgIf := range userIDsCfg.([]interface{}) {
			for _, id := range group.Users.ID {
				if id != idCfgIf.(int) {
					continue
				}
				appliedUserIDs = append(appliedUserIDs, id)
				break
			}
		}

		if len(appliedUserIDs) > 0 {
			err := d.Set("users", appliedUserIDs)
			if err != nil {
				return err
			}
		}
	}

	err = d.Set("admins", group.Admins.ID)
	if err != nil {
		return err
	}

	// TODO: Remove this part in release 0.6.0, this additional request is only
	// here to retrieve the quotas information
	groupInfo, err := controller.Group(group.ID).Info(false)
	if err != nil {
		return err
	}

	if _, ok := d.GetOk("quotas"); ok {
		err = flattenQuotasMapFromStructs(d, &groupInfo.QuotasList)
		if err != nil {
			return err
		}
	}

	return nil
}
