package opennebula

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"

	ver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable TLS validation",
				DefaultFunc: schema.EnvDefaultFunc("OPENNEBULA_INSECURE", false),
			},
			"default_tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Add default tags to the resources",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tags": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "Default tags to apply",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
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
			"opennebula_host":                  dataOpennebulaHost(),
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
			"opennebula_virtual_network_address_range":    resourceOpennebulaVirtualNetworkAddressRange(),
			"opennebula_cluster":                          resourceOpennebulaCluster(),
			"opennebula_host":                             resourceOpennebulaHost(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

type Configuration struct {
	OneVersion     *ver.Version
	Controller     *goca.Controller
	mutex          MutexKV
	defaultTags    map[string]interface{}
	oldDefaultTags map[string]interface{}
	newDefaultTags map[string]interface{}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	var diags diag.Diagnostics

	username, ok := d.GetOk("username")
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "username should be defined",
		})
		return nil, diags
	}

	password, ok := d.GetOk("password")
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "password should be defined",
		})
		return nil, diags
	}

	endpoint, ok := d.GetOk("endpoint")
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "endpoint should be defined",
		})
		return nil, diags
	}

	insecure := d.Get("insecure")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure.(bool)},
	}

	oneClient := goca.NewClient(goca.NewConfig(username.(string),
		password.(string),
		endpoint.(string)),
		&http.Client{Transport: tr})

	versionStr, err := goca.NewController(oneClient).SystemVersion()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get OpenNebula release number",
			Detail:   err.Error(),
		})
		return nil, diags
	}
	version, err := ver.NewVersion(versionStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse OpenNebula version",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	log.Printf("[INFO] OpenNebula version: %s", versionStr)

	cfg := &Configuration{
		OneVersion: version,
		mutex:      *NewMutexKV(),
	}

	defaultTagsOldIf, defaultTagsNewIf := d.GetChange("default_tags")
	defaultTagsOld := defaultTagsOldIf.(*schema.Set).List()
	defaultTagsNew := defaultTagsNewIf.(*schema.Set).List()

	defaultTags := d.Get("default_tags").(*schema.Set).List()
	if len(defaultTags) > 0 {
		defaultTagsMap := defaultTags[0].(map[string]interface{})
		cfg.defaultTags = defaultTagsMap["tags"].(map[string]interface{})
		if len(defaultTagsOld) > 0 {
			cfg.oldDefaultTags = defaultTagsOld[0].(map[string]interface{})
		}
		if len(defaultTagsNew) > 0 {
			cfg.newDefaultTags = defaultTagsNew[0].(map[string]interface{})
		}
	}

	flowEndpoint, ok := d.GetOk("flow_endpoint")
	if ok {
		flowClient := goca.NewDefaultFlowClient(
			goca.NewFlowConfig(username.(string),
				password.(string),
				flowEndpoint.(string)))

		cfg.Controller = goca.NewGenericController(oneClient, flowClient)
		return cfg, nil

	}

	cfg.Controller = goca.NewController(oneClient)

	return cfg, nil
}
