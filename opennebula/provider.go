package opennebula

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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
			"flow_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The URL to your public or private OpenNebula Flow server",
				DefaultFunc: schema.EnvDefaultFunc("OPENNEBULA_FLOW_ENDPOINT", nil),
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
			"opennebula_cluster":               dataOpennebulaCluster(),
			"opennebula_group":                 dataOpennebulaGroup(),
			"opennebula_image":                 dataOpennebulaImage(),
			"opennebula_security_group":        dataOpennebulaSecurityGroup(),
			"opennebula_template":              dataOpennebulaTemplate(),
			"opennebula_user":                  dataOpennebulaUser(),
			"opennebula_virtual_data_center":   dataOpennebulaVirtualDataCenter(),
			"opennebula_virtual_network":       dataOpennebulaVirtualNetwork(),
			"opennebula_virtual_machine_group": dataOpennebulaVMGroup(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"opennebula_acl":                              resourceOpennebulaACL(),
			"opennebula_group":                            resourceOpennebulaGroup(),
			"opennebula_group_admins":                     resourceOpennebulaGroupAdmins(),
			"opennebula_image":                            resourceOpennebulaImage(),
			"opennebula_security_group":                   resourceOpennebulaSecurityGroup(),
			"opennebula_template":                         resourceOpennebulaTemplate(),
			"opennebula_user":                             resourceOpennebulaUser(),
			"opennebula_virtual_data_center":              resourceOpennebulaVirtualDataCenter(),
			"opennebula_virtual_machine":                  resourceOpennebulaVirtualMachine(),
			"opennebula_virtual_network":                  resourceOpennebulaVirtualNetwork(),
			"opennebula_virtual_machine_group":            resourceOpennebulaVMGroup(),
			"opennebula_service":                          resourceOpennebulaService(),
			"opennebula_service_template":                 resourceOpennebulaServiceTemplate(),
			"opennebula_virtual_router_instance":          resourceOpennebulaVirtualRouterInstance(),
			"opennebula_virtual_router_instance_template": resourceOpennebulaVirtualRouterInstanceTemplate(),
			"opennebula_virtual_router":                   resourceOpennebulaVirtualRouter(),
			"opennebula_virtual_router_nic":               resourceOpennebulaVirtualRouterNIC(),
		},

		ConfigureFunc: providerConfigure,
	}
}

type Configuration struct {
	OneVersion string
	Controller *goca.Controller
	mutex      MutexKV
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	username, ok := d.GetOk("username")
	if !ok {
		return nil, fmt.Errorf("username should be defined")
	}

	password, ok := d.GetOk("password")
	if !ok {
		return nil, fmt.Errorf("password should be defined")
	}

	endpoint, ok := d.GetOk("endpoint")
	if !ok {
		return nil, fmt.Errorf("endpoint should be defined")
	}

	oneClient := goca.NewDefaultClient(goca.NewConfig(username.(string),
		password.(string),
		endpoint.(string)))

	version, err := goca.NewController(oneClient).SystemVersion()
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] OpenNebula version: %s", version)

	flowEndpoint, ok := d.GetOk("flow_endpoint")
	if ok {
		flowClient := goca.NewDefaultFlowClient(
			goca.NewFlowConfig(username.(string),
				password.(string),
				flowEndpoint.(string)))

		return &Configuration{
			OneVersion: version,
			Controller: goca.NewGenericController(oneClient, flowClient),
			mutex:      *NewMutexKV(),
		}, nil

	}

	return &Configuration{
		OneVersion: version,
		Controller: goca.NewController(oneClient),
		mutex:      *NewMutexKV(),
	}, nil
}
