package opennebula

import (
	"context"
	"fmt"
	"strconv"

	userSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaUserRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the User",
			},
			"primary_group": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Primary (Default) Group ID of the user. Defaults to 0",
			},
			"groups": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of group IDs to add to the user",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func userFilter(d *schema.ResourceData, meta interface{}) (*userSc.User, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	users, err := controller.Users().Info()
	if err != nil {
		return nil, err
	}

	// filter users with user defined criterias
	name, nameOk := d.GetOk("name")
	primaryGroup, primaryGroupOk := d.GetOk("primary_group")
	groups, groupsOk := d.GetOk("groups")
	tagsInterface, tagsOk := d.GetOk("tags")

	match := make([]*userSc.User, 0, 1)
userLoop:
	for i, user := range users.Users {

		if nameOk && user.Name != name {
			continue
		}

		if groupsOk {
			groupSet := make(map[int]struct{})
			for _, id := range user.Groups.ID {
				groupSet[id] = struct{}{}
			}
			for _, group := range groups.([]interface{}) {
				_, ok := groupSet[group.(int)]
				if !ok {
					continue userLoop
				}
			}
		}

		if primaryGroupOk && user.GID != primaryGroup {
			continue
		}

		if tagsOk {
			tags := tagsInterface.(map[string]interface{})

			if !matchTags(user.Template, tags) {
				continue
			}
		}

		match = append(match, &users.Users[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no user match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several users match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	user, err := userFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "users filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(user.Template)

	d.SetId(strconv.FormatInt(int64(user.ID), 10))
	d.Set("name", user.Name)
	d.Set("auth_driver", user.AuthDriver)
	d.Set("primary_group", user.GID)

	err = flattenUserGroups(d, user)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "failed to flatten groups",
			Detail:   fmt.Sprintf("User (ID: %d): %s", user.ID, err),
		})
		return diags
	}

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("User (ID: %d): %s", user.ID, err),
			})
			return diags
		}
	}

	return nil
}
