package opennebula

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataOpennebulaUser() *schema.Resource {
	return &schema.Resource{
		Read: resourceOpennebulaUserRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the User",
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
