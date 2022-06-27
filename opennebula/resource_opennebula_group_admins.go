package opennebula

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpennebulaGroupAdmins() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaGroupAdminsCreate,
		ReadContext:   resourceOpennebulaGroupAdminsRead,
		UpdateContext: resourceOpennebulaGroupAdminsUpdate,
		DeleteContext: resourceOpennebulaGroupAdminsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Name of the group",
			},
			"users_ids": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of user IDs admins of the group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func resourceOpennebulaGroupAdminsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)
	gc := controller.Group(groupID)

	// add admins users_ids if list provided
	adminsIDs := d.Get("users_ids").([]interface{})
	for _, id := range adminsIDs {
		err := gc.AddAdmin(id.(int))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to add an admin to the group",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	d.SetId(fmt.Sprint(groupID))

	return resourceOpennebulaGroupAdminsRead(ctx, d, meta)
}

func resourceOpennebulaGroupAdminsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the group admins ID",
			Detail:   err.Error(),
		})
		return diags
	}

	group, err := controller.Group(int(groupID)).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve group informations",
			Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
		})
		return diags
	}

	err = d.Set("users_ids", group.Admins.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set field",
			Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
		})
		return diags
	}

	return nil
}

func resourceOpennebulaGroupAdminsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the group admins ID",
			Detail:   err.Error(),
		})
		return diags
	}
	gc := controller.Group(int(groupID))

	oldUsersIf, newUsersIf := d.GetChange("users_ids")

	oldUsers := schema.NewSet(schema.HashInt, oldUsersIf.([]interface{}))
	newUsers := schema.NewSet(schema.HashInt, newUsersIf.([]interface{}))

	// delete admins
	remUsers := oldUsers.Difference(newUsers)

	for _, id := range remUsers.List() {
		err := gc.DelAdmin(id.(int))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to delete a group admin",
				Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
			})
			return diags
		}
	}

	// add admins
	addUsers := newUsers.Difference(oldUsers)

	for _, id := range addUsers.List() {
		err := gc.AddAdmin(id.(int))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to add a group admin",
				Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
			})
			return diags
		}
	}

	return resourceOpennebulaGroupAdminsRead(ctx, d, meta)
}

func resourceOpennebulaGroupAdminsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the group admins ID",
			Detail:   err.Error(),
		})
		return diags
	}
	gc := controller.Group(int(groupID))

	// add admins users_ids if list provided
	adminsIDs := d.Get("users_ids").([]interface{})
	for _, id := range adminsIDs {
		err := gc.DelAdmin(id.(int))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to delete a group admin",
				Detail:   fmt.Sprintf("group (ID: %d): %s", groupID, err),
			})
			return diags
		}
	}

	return nil
}
