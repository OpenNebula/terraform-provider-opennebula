package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/marketplace"
)

const (
	One             string = "one"
	Http                   = "http"
	S3                     = "s3"
	LinuxContainers        = "linuxcontainers"
	TurnkeyLinux           = "turnkeylinux"
	DockerHub              = "dockerhub"
)

var S3Types = []string{"aws", "minio", "ceph"}

var defaultMarketMinTimeout = 20
var defaultMarketTimeout = time.Duration(defaultHostMinTimeout) * time.Minute

// Common schema to linux containers LXC and Turnkey linux
func commonBackendSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"endpoint_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The base URL of the Market",
		},
		"roofs_image_size": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Size in MB for the image holding the rootfs",
		},
		"filesystem": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filesystem used for the image",
		},
		"image_block_file_format": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Image block file format",
		},
		"skip_untested": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Include only apps with support for context",
		},
	}
}

func resourceOpennebulaMarketPlace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaMarketPlaceCreate,
		ReadContext:   resourceOpennebulaMarketPlaceRead,
		UpdateContext: resourceOpennebulaMarketPlaceUpdate,
		DeleteContext: resourceOpennebulaMarketPlaceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultMarketTimeout),
			Update: schema.DefaultTimeout(defaultMarketTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the marketplace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the marketplace",
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the marketplace (in Unix format, owner-group-other, use-manage-admin)",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if len(value) != 3 {
						errors = append(errors, fmt.Errorf("%q has specify 3 permission sets: owner-group-other", k))
					}

					all := true
					for _, c := range strings.Split(value, "") {
						if c < "0" || c > "7" {
							all = false
						}
					}
					if !all {
						errors = append(errors, fmt.Errorf("Each character in %q should specify a Unix-like permission set with a number from 0 to 7", k))
					}

					return
				},
			},
			"uid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the user that will own the marketplace",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the marketplace",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the marketplace",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the marketplace",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allow to enable or disable the market place",
			},
			"one": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The marketplace endpoint url",
						},
					},
				},
				ConflictsWith: []string{"http", "s3", "lxc", "turnkey", "dockerhub"},
			},
			"http": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Base URL of the Marketplace HTTP endpoint",
						},
						"path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Absolute directory path to place images in the front-end or in the hosts pointed at by storage_bridge_list",
						},
						"storage_bridge_list": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "List of servers to access the public directory",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							MinItems: 1,
						},
					},
				},
				ConflictsWith: []string{"one", "s3", "lxc", "turnkey", "dockerhub"},
			},
			"s3": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Type of the s3 backend: aws, ceph, minio",
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {

								if !contains(v.(string), S3Types) {
									errors = append(errors, fmt.Errorf("s3 backend \"type\" must be one of: %s", strings.Join(S3Types, ",")))
								}

								return
							},
						},
						"access_key_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The access key of the S3 user",
						},
						"secret_access_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The secret key of the S3 user",
						},
						"bucket": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The bucket where the files will be stored",
						},
						"region": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The region to connect to. Any value will work with Ceph-S3",
						},
						"endpoint_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Only required when connecteing to a service other than Amazon S3",
						},
						"total_size": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     1048576,
							Description: "Define the total size of the marketplace in MB.",
						},
						"read_block_length": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     100,
							Description: "Split the file into chunks of this size in MB",
						},
					},
				},
				ConflictsWith: []string{"one", "http", "lxc", "turnkey", "dockerhub"},
			},
			"lxc": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: mergeSchemas(
						commonBackendSchema(),
						map[string]*schema.Schema{
							"cpu": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "VM template CPU",
							},
							"vcpu": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "VM template VCPU",
							},
							"memory": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "VM template memory",
							},
							"privileged": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Secrurity mode of the Linux Container",
							},
						}),
				},
				ConflictsWith: []string{"one", "http", "s3", "turnkey", "dockerhub"},
			},
			"turnkey": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: commonBackendSchema(),
				},
				ConflictsWith: []string{"one", "http", "s3", "lxc", "dockerhub"},
			},
			"dockerhub": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"one", "http", "s3", "lxc", "turnkey"},
			},
			"tags":             tagsSchema(),
			"default_tags":     defaultTagsSchemaComputed(),
			"tags_all":         tagsSchemaComputed(),
			"template_section": templateSectionSchema(),
		},
	}
}

func getMarketPlaceController(d *schema.ResourceData, meta interface{}) (*goca.MarketPlaceController, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	var gc *goca.MarketPlaceController

	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 0)
		if err != nil {
			return nil, err
		}
		gc = controller.MarketPlace(int(gid))
	}

	return gc, nil
}

func resourceOpennebulaMarketPlaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	tpl, diags := generateMarketplaceTemplate(d, meta)
	if len(diags) > 0 {
		return diags
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
	}

	// add default tags if they aren't overriden
	if len(config.defaultTags) > 0 {
		for k, v := range config.defaultTags {
			key := strings.ToUpper(k)
			p, _ := tpl.GetPair(key)
			if p != nil {
				continue
			}
			tpl.AddPair(key, v)
		}
	}

	log.Printf("[DEBUG] create marketplace with template: %s", tpl.String())

	marketplaceID, err := controller.MarketPlaces().Create(tpl.String())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the marketplace",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%d", marketplaceID))

	mpc := controller.MarketPlace(marketplaceID)

	// update permisions
	if perms, ok := d.GetOk("permissions"); ok {
		err = mpc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",
				Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	// manage enabled/disabled state
	disabled := d.Get("disabled").(bool)
	if disabled {
		err := mpc.Enable(!disabled)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to enable/disable the marketplace",
				Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		timeout := d.Timeout(schema.TimeoutCreate)
		_, err = waitForMarketplaceStates(ctx, mpc, timeout, []string{marketplace.Enabled.String()}, []string{marketplace.Disabled.String()})
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to wait marketplace to be in DISABLED state",
				Detail:   fmt.Sprintf("marketplace marketplace (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaMarketPlaceRead(ctx, d, meta)
}

func generateMarketplaceTemplate(d *schema.ResourceData, meta interface{}) (*marketplace.Template, diag.Diagnostics) {

	tpl := marketplace.NewTemplate()
	var diags diag.Diagnostics

	name, ok := d.GetOk("name")
	if ok {
		tpl.AddPair("NAME", name.(string))
	}

	description, ok := d.GetOk("description")
	if ok {
		tpl.AddPair("DESCRIPTION", description.(string))
	}

	diags = generateBackendOne(d, meta, tpl)
	if len(diags) > 0 {
		return nil, diags
	}

	diags = generateBackendHttp(d, meta, tpl)
	if len(diags) > 0 {
		return nil, diags
	}

	diags = generateBackendS3(d, meta, tpl)
	if len(diags) > 0 {
		return nil, diags
	}

	diags = generateBackendLXC(d, meta, tpl)
	if len(diags) > 0 {
		return nil, diags
	}

	diags = generateBackendTurnkey(d, meta, tpl)
	if len(diags) > 0 {
		return nil, diags
	}

	backendDockerhub := d.Get("dockerhub").(bool)
	if backendDockerhub {
		tpl.AddPair("MARKET_MAD", DockerHub)
	}

	return tpl, nil
}

func generateBackendOne(d *schema.ResourceData, meta interface{}, tpl *marketplace.Template) diag.Diagnostics {
	var diags diag.Diagnostics

	backendOneList := d.Get("one").(*schema.Set).List()
	if len(backendOneList) > 0 {

		backendOneIf := backendOneList[0]
		backendOne := backendOneIf.(map[string]interface{})

		tpl.AddPair("MARKET_MAD", One)

		endpoint := backendOne["endpoint_url"].(string)
		if len(endpoint) > 0 {
			tpl.AddPair("ENDPOINT", endpoint)

		}

	}

	return diags
}

func generateBackendHttp(d *schema.ResourceData, meta interface{}, tpl *marketplace.Template) diag.Diagnostics {
	var diags diag.Diagnostics

	backendHttpList := d.Get("http").(*schema.Set).List()
	if len(backendHttpList) > 0 {

		tpl.AddPair("MARKET_MAD", Http)

		backendHttpIf := backendHttpList[0]
		backendHttp := backendHttpIf.(map[string]interface{})

		endpoint := backendHttp["endpoint_url"].(string)
		tpl.AddPair("BASE_URL", endpoint)

		path := backendHttp["path"].(string)
		tpl.AddPair("PUBLIC_DIR", path)

		bridgeList := backendHttp["storage_bridge_list"].(*schema.Set).List()
		if len(bridgeList) > 0 {
			bridgeListStr := bridgeList[0].(string)
			for _, bridge := range bridgeList[1:] {
				bridgeListStr += " " + bridge.(string)
			}
			tpl.AddPair("BRIDGE_LIST", bridgeListStr)
		}

	}

	return diags
}

func generateBackendS3(d *schema.ResourceData, meta interface{}, tpl *marketplace.Template) diag.Diagnostics {
	var diags diag.Diagnostics

	backendS3List := d.Get("s3").(*schema.Set).List()
	if len(backendS3List) > 0 {

		backendS3If := backendS3List[0]
		backendS3 := backendS3If.(map[string]interface{})

		tpl.AddPair("MARKET_MAD", S3)

		accessKeyID := backendS3["access_key_id"].(string)
		tpl.AddPair("ACCESS_KEY_ID", accessKeyID)

		secretAccessKey := backendS3["secret_access_key"].(string)
		tpl.AddPair("SECRET_ACCESS_KEY", secretAccessKey)

		bucket := backendS3["bucket"].(string)
		tpl.AddPair("BUCKET", bucket)

		region := backendS3["region"].(string)
		tpl.AddPair("REGION", region)

		s3BackendType, ok := backendS3["type"]
		if ok {
			switch s3BackendType.(string) {
			case "aws":
				// AWS, SIGNATURE_VERSION, FORCE_PATH_STYLE are left blank
			case "ceph":
				tpl.AddPair("SIGNATURE_VERSION", "s3")
				tpl.AddPair("FORCE_PATH_STYLE", "YES")
				tpl.AddPair("AWS", "no")
			default:
				tpl.AddPair("AWS", "no")
			}
		}

		endpoint := backendS3["endpoint_url"].(string)
		if len(endpoint) > 0 {
			tpl.AddPair("ENDPOINT", endpoint)
		}

		totalMB := backendS3["total_size"].(int)
		tpl.AddPair("TOTAL_MB", totalMB)

		readLen := backendS3["read_block_length"].(int)
		tpl.AddPair("READ_LENGTH", readLen)

	}

	return diags
}

func generateBackendCommon(backendMap map[string]interface{}, tpl *marketplace.Template) {

	endpoint := backendMap["endpoint_url"].(string)
	if len(endpoint) > 0 {
		tpl.AddPair("ENDPOINT", endpoint)
	}

	rootFSImageSize := backendMap["roofs_image_size"].(int)
	if rootFSImageSize > 0 {
		tpl.AddPair("IMAGE_SIZE_MB", rootFSImageSize)
	}

	filesystem := backendMap["filesystem"].(string)
	if len(filesystem) > 0 {
		tpl.AddPair("FILESYSTEM", filesystem)
	}

	imageBlockFileFormat := backendMap["image_block_file_format"].(string)
	if len(imageBlockFileFormat) > 0 {
		tpl.AddPair("FORMAT", imageBlockFileFormat)
	}

	skipUntested := backendMap["skip_untested"].(bool)
	if skipUntested {
		tpl.AddPair("SKIP_UNTESTED", skipUntested)
	}
}

func generateBackendLXC(d *schema.ResourceData, meta interface{}, tpl *marketplace.Template) diag.Diagnostics {
	var diags diag.Diagnostics

	backendLXCList := d.Get("lxc").(*schema.Set).List()
	if len(backendLXCList) > 0 {

		backendLXCIf := backendLXCList[0]
		backendLXC := backendLXCIf.(map[string]interface{})

		tpl.AddPair("MARKET_MAD", LinuxContainers)

		generateBackendCommon(backendLXC, tpl)

		CPU := backendLXC["cpu"].(int)
		if CPU > 0 {
			tpl.AddPair("CPU", CPU)
		}

		vCPU := backendLXC["vcpu"].(int)
		if vCPU > 0 {
			tpl.AddPair("VCPU", vCPU)
		}

		memory := backendLXC["memory"].(int)
		if memory > 0 {
			tpl.AddPair("MEMORY", memory)
		}

		privileged, ok := backendLXC["privileged"]
		if ok {
			tpl.AddPair("PRIVILEGED", privileged)
		}

	}

	return diags
}

func generateBackendTurnkey(d *schema.ResourceData, meta interface{}, tpl *marketplace.Template) diag.Diagnostics {
	var diags diag.Diagnostics

	backendTurnkeyList := d.Get("turnkey").(*schema.Set).List()
	if len(backendTurnkeyList) > 0 {

		backendTurnkeyIf := backendTurnkeyList[0]
		backendTurnkey := backendTurnkeyIf.(map[string]interface{})

		tpl.AddPair("MARKET_MAD", TurnkeyLinux)

		generateBackendCommon(backendTurnkey, tpl)

	}

	return diags
}

func resourceOpennebulaMarketPlaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	mpc, err := getMarketPlaceController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the marketplace controller",
			Detail:   err.Error(),
		})
		return diags
	}

	marketplaceInfos, err := mpc.Info(false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing marketplace %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed retrieve marketplace informations",
			Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.Set("name", marketplaceInfos.Name)
	d.Set("uid", marketplaceInfos.UID)
	d.Set("gid", marketplaceInfos.GID)
	d.Set("uname", marketplaceInfos.UName)
	d.Set("gname", marketplaceInfos.GName)
	d.Set("permissions", permissionsUnixString(*marketplaceInfos.Permissions))

	description, err := marketplaceInfos.Template.GetStr("DESCRIPTION")
	if err == nil {
		d.Set("description", description)
	}

	switch marketplaceInfos.MarketMad {
	case One:
		endpointUrl, _ := marketplaceInfos.Template.GetStr("ENDPOINT")
		d.Set("one", []map[string]interface{}{
			{
				"endpoint_url": endpointUrl,
			},
		})
	case Http:
		publicDir, _ := marketplaceInfos.Template.GetStr("PUBLIC_DIR")
		baseUrl, _ := marketplaceInfos.Template.GetStr("BASE_URL")

		backendHttp := map[string]interface{}{
			"endpoint_url": baseUrl,
			"path":         publicDir,
		}

		bridgeListStr, err := marketplaceInfos.Template.GetStr("BRIDGE_LIST")
		if err == nil {
			bridgeList := strings.Split(bridgeListStr, " ")
			backendHttp["storage_bridge_list"] = bridgeList
		}

		d.Set("http", []map[string]interface{}{backendHttp})
	case S3:

		accessKeyID, _ := marketplaceInfos.Template.GetStr("ACCESS_KEY_ID")
		secretAccessKey, _ := marketplaceInfos.Template.GetStr("SECRET_ACCESS_KEY")
		bucket, _ := marketplaceInfos.Template.GetStr("BUCKET")
		region, _ := marketplaceInfos.Template.GetStr("REGION")

		backendS3 := map[string]interface{}{
			"access_key_id":     accessKeyID,
			"secret_access_key": secretAccessKey,
			"bucket":            bucket,
			"region":            region,
		}

		endpoint, err := marketplaceInfos.Template.GetStr("ENDPOINT")
		if err == nil {
			backendS3["endpoint_url"] = endpoint
		}

		totalMB, err := marketplaceInfos.Template.GetI("TOTAL_MB")
		if err == nil {
			backendS3["total_size"] = totalMB
		}

		readLength, err := marketplaceInfos.Template.GetI("READ_LENGTH")
		if err == nil {
			backendS3["read_block_length"] = readLength
		}

		aws, err := marketplaceInfos.Template.GetStr("AWS")
		if err != nil || len(aws) == 0 {
			// no tags or empty type means AWS type
			backendS3["type"] = "aws"
		} else {
			sigVersion, err := marketplaceInfos.Template.GetStr("SIGNATURE_VERSION")
			if err == nil && sigVersion == "s3" {
				backendS3["type"] = "ceph"
			} else {
				backendS3["type"] = "minio"
			}
		}

		d.Set("s3", []map[string]interface{}{backendS3})

	case LinuxContainers:
		backendLXC := flattenCommonBackend(d, meta, &marketplaceInfos.Template)

		cpu, err := marketplaceInfos.Template.GetInt("CPU")
		if err == nil {
			backendLXC["cpu"] = cpu
		}

		vcpu, err := marketplaceInfos.Template.GetInt("VCPU")
		if err == nil {
			backendLXC["vcpu"] = vcpu
		}

		mem, err := marketplaceInfos.Template.GetInt("MEMORY")
		if err == nil {
			backendLXC["memory"] = mem
		}

		privileged, err := marketplaceInfos.Template.GetStr("PRIVILEGED")
		if err == nil {
			backendLXC["privileged"] = privileged
		}
		d.Set("lxc", []map[string]interface{}{backendLXC})
	case TurnkeyLinux:

		backendTurnkey := flattenCommonBackend(d, meta, &marketplaceInfos.Template)

		d.Set("turnkey", []map[string]interface{}{backendTurnkey})

	case DockerHub:
		d.Set("dockerhub", true)
	default:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The marketplace backend type is not handled by the provider",
			Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	flattenDiags := flattenMarketplaceTemplate(d, meta, &marketplaceInfos.Template)
	for _, diag := range flattenDiags {
		diags = append(diags, diag)
	}

	state, _ := marketplaceInfos.StateString()
	d.Set("disabled", state == marketplace.Disabled.String())

	return diags
}

// Common schema to linux containers LXC and Turnkey linux
func flattenCommonBackend(d *schema.ResourceData, meta interface{}, marketplaceTpl *marketplace.Template) map[string]interface{} {
	backend := make(map[string]interface{})

	endpoint, err := marketplaceTpl.GetStr("ENDPOINT")
	if err == nil {
		backend["endpoint_url"] = endpoint
	}

	imageSizeMB, err := marketplaceTpl.GetInt("IMAGE_SIZE_MB")
	if err == nil {
		backend["roofs_image_size"] = imageSizeMB
	}

	filesystem, err := marketplaceTpl.GetStr("FILESYSTEM")
	if err == nil {
		backend["filesystem"] = filesystem
	}

	format, err := marketplaceTpl.GetStr("FORMAT")
	if err == nil {
		backend["image_block_file_format"] = format
	}

	skipUntested, err := marketplaceTpl.GetStr("SKIP_UNTESTED")
	if err == nil {
		backend["skip_untested"] = skipUntested
	}

	return backend
}

func flattenMarketplaceTemplate(d *schema.ResourceData, meta interface{}, marketplaceTpl *marketplace.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	err := flattenTemplateSection(d, meta, &marketplaceTpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to read template section",
			Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
		})
	}

	flattenDiags := flattenTemplateTags(d, meta, &marketplaceTpl.Template)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
	}

	return diags
}

func resourceOpennebulaMarketPlaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	//config := meta.(*Configuration)
	//controller := config.Controller

	var diags diag.Diagnostics

	mpc, err := getMarketPlaceController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the marketplace controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// template management

	marketplaceInfos, err := mpc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		err := mpc.Rename(newName)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	update := false
	newTpl := marketplaceInfos.Template

	if d.HasChange("one") {
		newTpl.Del("MARKET_MAD")
		newTpl.Del("ENDPOINT")

		diags = generateBackendOne(d, meta, &newTpl)
		if len(diags) > 0 {
			return diags
		}

		update = true
	}

	if d.HasChange("http") {
		newTpl.Del("MARKET_MAD")
		newTpl.Del("BASE_URL")
		newTpl.Del("PUBLIC_DIR")
		newTpl.Del("BRIDGE_LIST")

		diags = generateBackendHttp(d, meta, &newTpl)
		if len(diags) > 0 {
			return diags
		}

		update = true
	}

	if d.HasChange("s3") {

		for _, k := range []string{"MARKET_MAD", "ACCESS_KEY_ID", "SECRET_ACCESS_KEY",
			"BUCKET", "REGION", "ENDPOINT",
			"SIGNATURE_VERSION", "FORCE_PATH_STYLE", "TOTAL_MB",
			"READ_LENGTH", "AWS"} {
			newTpl.Del(k)
		}

		diags = generateBackendS3(d, meta, &newTpl)
		if len(diags) > 0 {
			return diags
		}

		update = true
	}

	if d.HasChange("lxc") {

		for _, k := range []string{"MARKET_MAD", "ENDPOINT", "IMAGE_SIZE_MB", "FILESYSTEM",
			"FORMAT", "SKIP_UNTESTED", "CPU", "VCPU", "MEMORY", "PRIVILEGED"} {
			newTpl.Del(k)
		}

		diags = generateBackendLXC(d, meta, &newTpl)
		if len(diags) > 0 {
			return diags
		}

		update = true
	}

	if d.HasChange("template_section") {

		updateTemplateSection(d, &newTpl.Template)

		update = true
	}

	if d.HasChange("tags") {

		oldTagsIf, newTagsIf := d.GetChange("tags")
		oldTags := oldTagsIf.(map[string]interface{})
		newTags := newTagsIf.(map[string]interface{})

		// delete tags
		for k, _ := range oldTags {
			_, ok := newTags[k]
			if ok {
				continue
			}
			newTpl.Del(strings.ToUpper(k))
		}

		// add/update tags
		for k, v := range newTags {
			key := strings.ToUpper(k)
			newTpl.Del(key)
			newTpl.AddPair(key, v)
		}

		update = true
	}

	if d.HasChange("tags_all") {
		oldTagsAllIf, newTagsAllIf := d.GetChange("tags_all")
		oldTagsAll := oldTagsAllIf.(map[string]interface{})
		newTagsAll := newTagsAllIf.(map[string]interface{})

		tags := d.Get("tags").(map[string]interface{})

		// delete tags
		for k, _ := range oldTagsAll {
			_, ok := newTagsAll[k]
			if ok {
				continue
			}
			newTpl.Del(strings.ToUpper(k))
		}

		// reapply all default tags that were neither applied nor overriden via tags section
		for k, v := range newTagsAll {
			_, ok := tags[k]
			if ok {
				continue
			}

			key := strings.ToUpper(k)
			newTpl.Del(key)
			newTpl.AddPair(key, v)
		}

		update = true
	}

	if update {
		log.Printf("[DEBUG] update marketplace template: %s", newTpl.String())
		err = mpc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update marketplace content",
				Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

	}

	if d.HasChange("disabled") {
		disabled := d.Get("disabled").(bool)
		err := mpc.Enable(!disabled)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to enable/disable the marketplace",
				Detail:   fmt.Sprintf("marketplace (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		// wait on state transition
		timeout := d.Timeout(schema.TimeoutUpdate)

		// expected state when disabling
		pendingStates := []string{marketplace.Enabled.String()}
		targetStates := []string{marketplace.Disabled.String()}
		// expected states when enabling
		if disabled {
			tmp := pendingStates
			pendingStates = targetStates
			targetStates = tmp
		}

		_, err = waitForMarketplaceStates(ctx, mpc, timeout, pendingStates, targetStates)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to wait marketplace to be in %s state", strings.Join(targetStates, ", ")),
				Detail:   fmt.Sprintf("marketplace marketplace (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaMarketPlaceRead(ctx, d, meta)
}

func resourceOpennebulaMarketPlaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	mpc, err := getMarketPlaceController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the marketplace controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = mpc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("marketplace (ID: %d): %s", mpc.ID, err),
		})
		return diags
	}

	return nil
}

// waitForMarketStates wait for a an marketplace to reach some expected states
func waitForMarketplaceStates(ctx context.Context, mpc *goca.MarketPlaceController, timeout time.Duration, pending, target []string) (interface{}, error) {

	stateChangeConf := resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
		Refresh: func() (interface{}, string, error) {

			log.Println("Refreshing marketplace state...")

			marketInfos, err := mpc.Info(false)
			if err != nil {
				if NoExists(err) {
					return marketInfos, "notfound", nil
				}
				return marketInfos, "", err
			}
			state, _ := marketInfos.StateString()

			log.Printf("Marketplace (ID:%d, name:%s) is currently in state %s", marketInfos.ID, marketInfos.Name, state)

			return marketInfos, state, nil
		},
	}

	return stateChangeConf.WaitForStateContext(ctx)

}
