package opennebula

import (
	"context"
	"fmt"
	"strconv"

	groupSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/group"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaGroupRead,

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
			"tags": tagsSchema(),
		},
	}
}

func groupFilter(d *schema.ResourceData, meta interface{}) (*groupSc.GroupShort, error) {

	config := meta.(*Configuration)
	controller := config.Controller

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
		return nil, fmt.Errorf("no group match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several groups match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	group, err := groupFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "groups filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(group.Template)

	d.SetId(strconv.FormatInt(int64(group.ID), 10))
	d.Set("name", group.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("Group (ID: %d): %s", group.ID, err),
			})
			return diags
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
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "setting attribute failed",
					Detail:   fmt.Sprintf("Group (ID: %d): %s", group.ID, err),
				})
				return diags
			}
		}
	}

	err = d.Set("admins", group.Admins.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "setting attribute failed",
			Detail:   fmt.Sprintf("Group (ID: %d): %s", group.ID, err),
		})
		return diags
	}

	return nil
}
