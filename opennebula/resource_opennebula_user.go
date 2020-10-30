package opennebula

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

var authTypes = []string{"core", "public", "ssh", "x509", "ldap", "server_cipher", "server_x509", "custom"}

func resourceOpennebulaUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaUserCreate,
		Read:   resourceOpennebulaUserRead,
		Update: resourceOpennebulaUserUpdate,
		Delete: resourceOpennebulaUserDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the User",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Password of the User. Required for all `auth_driver` options excepted 'ldap'",
			},
			"auth_driver": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "core",
				Description: "Authentication driver. Select between: core, public, ssh, x509, ldap, server_cipher, server_x509 and custom. Defaults to 'core'.",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if inArray(value, authTypes) < 0 {
						errors = append(errors, fmt.Errorf("Auth driver %q must be one of: %s", k, strings.Join(locktypes, ",")))
					}

					return
				},
			},
			"primary_group": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
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
			"quotas": quotasSchema(),
		},
	}
}

func getUserController(d *schema.ResourceData, meta interface{}) (*goca.UserController, error) {
	controller := meta.(*goca.Controller)
	var uc *goca.UserController

	// Try to find the User by ID, if specified
	if d.Id() != "" {
		uid, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		uc = controller.User(int(uid))
	}

	// Otherwise, try to find the User by name as the de facto compound primary key
	if d.Id() == "" {
		uid, err := controller.Users().ByName(d.Get("name").(string))
		if err != nil {
			return nil, err
		}
		uc = controller.User(uid)
	}

	return uc, nil
}

func resourceOpennebulaUserCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	userName := d.Get("name").(string)
	userAuthDriver := d.Get("auth_driver").(string)
	var userPassword string
	if userAuthDriver != "ldap" {
		userPassword_interface, ok := d.GetOk("password")
		if !ok {
			return fmt.Errorf("Password cannot be empty if auth_driver is: %s", userAuthDriver)
		}
		userPassword = userPassword_interface.(string)
	}
	userGroupLists := d.Get("groups").([]interface{})
	userGroups := make([]int, 0, 1+len(userGroupLists))

	// Start Group array with Primary group
	userGroups = append(userGroups, d.Get("primary_group").(int))

	// add groups to user if list provided
	for _, gid := range userGroupLists {
		userGroups = append(userGroups, gid.(int))
	}

	userID, err := controller.Users().Create(userName, userPassword, userAuthDriver, userGroups)
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%v", userID))

	uc := controller.User(userID)

	if _, ok := d.GetOk("quotas"); ok {
		err = uc.Quota(generateQuotas(d))
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaUserRead(d, meta)
}

func resourceOpennebulaUserRead(d *schema.ResourceData, meta interface{}) error {
	uc, err := getUserController(d, meta)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing user %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		return err
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	user, err := uc.Info(false)
	if err != nil {
		return err
	}

	d.SetId(strconv.FormatUint(uint64(user.ID), 10))
	d.Set("name", user.Name)
	sum := sha256.Sum256([]byte(d.Get("password").(string)))
	if fmt.Sprintf("%x", sum) == user.Password {
		d.Set("password", d.Get("password").(string))
	} else {
		return fmt.Errorf("password doesn't match")
	}
	d.Set("auth_driver", user.AuthDriver)
	d.Set("primary_group", user.GID)
	userGroups := make([]int, 0)
	for _, u := range user.Groups.ID {
		if u == user.GID {
			continue
		}
		userGroups = append(userGroups, u)
	}
	if len(userGroups) > 0 {
		err = d.Set("groups", userGroups)
		if err != nil {
			return err
		}
	}
	err = flattenQuotasMapFromStructs(d, &user.QuotasList)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaUserUpdate(d *schema.ResourceData, meta interface{}) error {
	uc, err := getUserController(d, meta)
	if err != nil {
		return err
	}

	if d.HasChange("password") {
		// update password
		err = uc.Passwd(d.Get("password").(string))
		if err != nil {
			return err
		}
	}

	if d.HasChange("auth_driver") {
		// Erase previous authentication driver, let password unchanged
		err = uc.Chauth(d.Get("auth_driver").(string), "")
		if err != nil {
			return err
		}
	}

	if d.HasChange("primary_group") {
		// change the primary group of the User
		err = uc.Chgrp(d.Get("primary_group").(int))
		if err != nil {
			return err
		}
	}

	if d.HasChange("groups") {
		// Update secondary group list
		oGroupsInterface, nGroupsInterface := d.GetChange("groups")
		oGroups := oGroupsInterface.([]interface{})
		nGroups := nGroupsInterface.([]interface{})
		for _, g := range oGroups {
			if g.(int) == d.Get("primary_group").(int) {
				continue
			}
			err = uc.DelGroup(g.(int))
			if err != nil {
				return err
			}
		}
		for _, g := range nGroups {
			if g.(int) == d.Get("primary_group").(int) {
				continue
			}
			err = uc.AddGroup(g.(int))
			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("quotas") {
		if _, ok := d.GetOk("quotas"); ok {
			err = uc.Quota(generateQuotas(d))
			if err != nil {
				return err
			}
		}
	}

	return resourceOpennebulaUserRead(d, meta)
}

func resourceOpennebulaUserDelete(d *schema.ResourceData, meta interface{}) error {
	gc, err := getUserController(d, meta)
	if err != nil {
		return err
	}

	err = gc.Delete()
	if err != nil {
		return err
	}

	return nil
}
