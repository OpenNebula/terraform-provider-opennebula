package opennebula

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strconv"

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
				Optional:    true,
				Description: "The URL to your public or private OpenNebula",
			},
			"flow_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The URL to your public or private OpenNebula Flow server",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the user to identify as",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The password for the user",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable TLS validation",
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
			"opennebula_templates":             dataOpennebulaTemplates(),
			"opennebula_user":                  dataOpennebulaUser(),
			"opennebula_virtual_data_center":   dataOpennebulaVirtualDataCenter(),
			"opennebula_virtual_network":       dataOpennebulaVirtualNetwork(),
			"opennebula_virtual_machine_group": dataOpennebulaVMGroup(),
			"opennebula_host":                  dataOpennebulaHost(),
			"opennebula_datastore":             dataOpennebulaDatastore(),
			"opennebula_zone":                  dataOpennebulaZone(),
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
			"opennebula_datastore":                        resourceOpennebulaDatastore(),
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

	username := d.Get("username").(string)
	if len(username) == 0 {
		username = os.Getenv("OPENNEBULA_USERNAME")
	}
	if len(username) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "username should be defined",
			Detail:   "username should be provided either via the configuration or via the OPENNEBULA_USERNAME environment variable",
		})
		return nil, diags
	}

	password := d.Get("password").(string)
	if len(password) == 0 {
		password = os.Getenv("OPENNEBULA_PASSWORD")
	}
	if len(password) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "password should be defined",
			Detail:   "password should be provided either via the configuration or via the OPENNEBULA_PASSWORD environment variable",
		})
		return nil, diags
	}

	endpoint := d.Get("endpoint").(string)
	if len(endpoint) == 0 {
		endpoint = os.Getenv("OPENNEBULA_ENDPOINT")
	}
	if len(endpoint) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "endpoint should be defined",
			Detail:   "endpoint should be provided either via the configuration or via the OPENNEBULA_ENDPOINT environment variable",
		})
		return nil, diags
	}

	insecureIf := d.Get("insecure")
	insecure := false
	if insecureIf == nil || !insecureIf.(bool) {
		insecureStr := os.Getenv("OPENNEBULA_INSECURE")

		var err error
		if len(insecureStr) > 0 {
			insecure, err = strconv.ParseBool(insecureStr)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to parse boolean value from the OPENNEBULA_INSECURE environment variable",
					Detail:   err.Error(),
				})
				return nil, diags
			}
		}
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	oneClient := goca.NewClient(goca.NewConfig(username,
		password,
		endpoint),
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

	flowEndpoint := d.Get("flow_endpoint").(string)
	if len(flowEndpoint) == 0 {
		flowEndpoint = os.Getenv("OPENNEBULA_FLOW_ENDPOINT")
	}

	if len(flowEndpoint) > 0 {

		flowClient := goca.NewDefaultFlowClient(
			goca.NewFlowConfig(username,
				password,
				flowEndpoint))

		cfg.Controller = goca.NewGenericController(oneClient, flowClient)
		return cfg, nil

	}

	cfg.Controller = goca.NewController(oneClient)

	return cfg, nil
}
