package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL to your public or private OpenNebula",
				DefaultFunc: schema.EnvDefaultFunc("OPENNEBULA_ENDPOINT", nil),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the user to identify as",
				DefaultFunc: schema.EnvDefaultFunc("OPENNEBULA_USERNAME", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The password for the user",
				DefaultFunc: schema.EnvDefaultFunc("OPENNEBULA_PASSWORD", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"opennebula_group":                 dataOpennebulaGroup(),
			"opennebula_image":                 dataOpennebulaImage(),
			"opennebula_security_group":        dataOpennebulaSecurityGroup(),
			"opennebula_template":              dataOpennebulaTemplate(),
			"opennebula_virtual_data_center":   dataOpennebulaVirtualDataCenter(),
			"opennebula_virtual_network":       dataOpennebulaVirtualNetwork(),
			"opennebula_virtual_machine_group": dataOpennebulaVMGroup(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"opennebula_acl":                   resourceOpennebulaACL(),
			"opennebula_group":                 resourceOpennebulaGroup(),
			"opennebula_image":                 resourceOpennebulaImage(),
			"opennebula_security_group":        resourceOpennebulaSecurityGroup(),
			"opennebula_template":              resourceOpennebulaTemplate(),
			"opennebula_virtual_data_center":   resourceOpennebulaVirtualDataCenter(),
			"opennebula_virtual_machine":       resourceOpennebulaVirtualMachine(),
			"opennebula_virtual_network":       resourceOpennebulaVirtualNetwork(),
			"opennebula_virtual_machine_group": resourceOpennebulaVMGroup(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := goca.NewDefaultClient(goca.NewConfig(d.Get("username").(string),
		d.Get("password").(string),
		d.Get("endpoint").(string)))

	return goca.NewController(client), nil
}
