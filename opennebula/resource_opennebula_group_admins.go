package opennebula

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpennebulaGroupAdmins() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaGroupAdminsCreate,
		Read:   resourceOpennebulaGroupAdminsRead,
		Update: resourceOpennebulaGroupAdminsUpdate,
		Delete: resourceOpennebulaGroupAdminsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceOpennebulaGroupAdminsCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller

	groupID := d.Get("group_id").(int)
	gc := controller.Group(groupID)

	// add admins users_ids if list provided
	adminsIDs := d.Get("users_ids").([]interface{})
	for _, id := range adminsIDs {
		err := gc.AddAdmin(id.(int))
		if err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprint(groupID))

	return resourceOpennebulaGroupAdminsRead(d, meta)
}

func resourceOpennebulaGroupAdminsRead(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*Configuration)
	controller := config.Controller

	groupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	group, err := controller.Group(int(groupID)).Info(false)
	if err != nil {
		return err
	}

	err = d.Set("users_ids", group.Admins.ID)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaGroupAdminsUpdate(d *schema.ResourceData, meta interface{}) error {

	config := meta.(*Configuration)
	controller := config.Controller

	groupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
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
			return err
		}
	}

	// add admins
	addUsers := newUsers.Difference(oldUsers)

	for _, id := range addUsers.List() {
		err := gc.AddAdmin(id.(int))
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaGroupAdminsRead(d, meta)
}

func resourceOpennebulaGroupAdminsDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller

	groupID, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	gc := controller.Group(int(groupID))

	// add admins users_ids if list provided
	adminsIDs := d.Get("users_ids").([]interface{})
	for _, id := range adminsIDs {
		err := gc.DelAdmin(id.(int))
		if err != nil {
			return err
		}
	}

	return nil
}
