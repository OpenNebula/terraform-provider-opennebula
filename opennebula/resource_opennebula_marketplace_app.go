package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"

	app "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/marketplaceapp"
	appk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/marketplaceapp/keys"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

var marketplaceAppType = []string{AppTypeImage, AppTypeVM, AppTypeService}
var marketAppPairingKey = "TMP_TF_RESOURCE_ID"

var defaultMarketAppMinTimeout = 20
var defaultMarketAppTimeout = time.Duration(defaultHostMinTimeout) * time.Minute

func resourceOpennebulaMarketPlaceApp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaMarketPlaceAppCreate,
		ReadContext:   resourceOpennebulaMarketPlaceAppRead,
		UpdateContext: resourceOpennebulaMarketPlaceAppUpdate,
		DeleteContext: resourceOpennebulaMarketPlaceAppDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultMarketAppTimeout),
			Update: schema.DefaultTimeout(defaultMarketAppTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"market_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the market to host the appliance",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the appliance",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of the app: IMAGE, VMTEMPLATE, SERVICE_TEMPLATE",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if !contains(value, marketplaceAppType) {
						errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(marketplaceAppType, ",")))
					}

					return
				},
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the appliance (in Unix format, owner-group-other, use-manage-admin)",
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
			"origin_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The ID of the source image",
				Default:     -1,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Text description of the appliance",
			},
			"publisher": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Publisher of the appliance",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A string indicating the appliance version",
			},
			"vmtemplate64": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Creates this template pointing to the base image",
			},
			"apptemplate64": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Associated template that will be added to the registered object",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the group owning the appliance",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allow to enable or disable the appliance",
			},
			"lock":             lockSchema(),
			"tags":             tagsSchema(),
			"default_tags":     defaultTagsSchemaComputed(),
			"tags_all":         tagsSchemaComputed(),
			"template_section": templateSectionSchema(),
		},
	}
}

func getMarketPlaceAppController(d *schema.ResourceData, meta interface{}) (*goca.MarketPlaceAppController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	appID, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.MarketPlaceApp(int(appID)), nil
}

func resourceOpennebulaMarketPlaceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	tpl := app.NewTemplate()

	if val, ok := d.GetOk("name"); ok {
		tpl.Add(appk.Name, val.(string))
	}
	if val, ok := d.GetOk("origin_id"); ok {
		tpl.Add(appk.OriginID, val.(int))
	}
	if val, ok := d.GetOk("type"); ok {
		tpl.Add(appk.Type, val.(string))
	}
	if val, ok := d.GetOk("description"); ok {
		tpl.Add(appk.Description, val.(string))
	}
	if val, ok := d.GetOk("publisher"); ok {
		tpl.Add(appk.Publisher, val.(string))
	}
	if val, ok := d.GetOk("version"); ok {
		tpl.Add(appk.Version, val.(string))
	}
	if val, ok := d.GetOk("vmtemplate64"); ok {
		tpl.Add(appk.VMTemplate64, val.(string))
	}
	if val, ok := d.GetOk("apptemplate64"); ok {
		tpl.Add(appk.AppTemplate64, val.(string))
	}

	vectorsInterface := d.Get("template_section").(*schema.Set).List()
	if len(vectorsInterface) > 0 {
		addTemplateVectors(vectorsInterface, &tpl.Template)
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

	log.Printf("[DEBUG] create marketplace appliance with template: %s", tpl.String())

	// allow resource identification in case of an error
	tmpProviderID, err := uuid.GenerateUUID()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate an temporary ID used to identify the new marketplace appliance",
			Detail:   err.Error(),
		})
		return diags
	}
	tpl.AddPair(marketAppPairingKey, tmpProviderID)

	// create the appliance
	marketID := d.Get("market_id").(int)
	appID, creationErr := controller.MarketPlaceApps().Create(tpl.String(), marketID)

	if appID != -1 {
		d.SetId(fmt.Sprintf("%d", appID))
	} else {
		// In case of an error, before returning, retrieve the appliance from the pool via the temporary ID added just above
		appPool, err := controller.MarketPlaceApps().Info()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve marketplace appliances pool",
				Detail:   err.Error(),
			})
			return diags
		}
		for _, app := range appPool.MarketPlaceApps {
			pairingID, _ := app.Template.GetStr(marketAppPairingKey)

			if pairingID == tmpProviderID {
				d.SetId(fmt.Sprintf("%d", app.ID))
				break
			}
		}

	}

	// check appliance creation error
	if creationErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the appliance",
			Detail:   creationErr.Error(),
		})
		return diags
	}

	log.Printf("[INFO] Market place appliance created")

	ac := controller.MarketPlaceApp(appID)

	timeout := d.Timeout(schema.TimeoutCreate)
	_, err = waitForMarketAppStates(ctx, ac, timeout, []string{app.Init.String(), app.Ready.String()}, []string{app.Ready.String()})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to wait appliance to be in READY state",
			Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// update permissions
	if perms, ok := d.GetOk("permissions"); ok {
		err = ac.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	// remove temporary pairing key
	appInfos, err := ac.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve appliance informations",
			Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	appInfos.Template.Del(marketAppPairingKey)

	err = ac.Update(appInfos.Template.String(), parameters.Replace)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to update appliance template",
			Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// manage enabled/disabled state
	disabled := d.Get("disabled").(bool)
	if disabled {
		err := ac.Enable(!disabled)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create the appliance",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		_, err = waitForMarketAppStates(ctx, ac, timeout, []string{app.Ready.String()}, []string{app.Disabled.String()})
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to wait appliance to be in DISABLED state",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	// manage appliance locking
	if lock, ok := d.GetOk("lock"); ok && lock.(string) != "UNLOCK" {

		var level shared.LockLevel
		err := StringToLockLevel(lock.(string), &level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to convert lock level",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = ac.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaMarketPlaceAppRead(ctx, d, meta)
}

func resourceOpennebulaMarketPlaceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	ac, err := getMarketPlaceAppController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the appliance controller",
			Detail:   err.Error(),
		})
		return diags
	}

	appInfos, err := ac.Info(false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing appliance %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed retrieve appliance informations",
			Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.Set("name", appInfos.Name)
	d.Set("origin_id", appInfos.OriginID)
	d.Set("type", ApplianceTypeToString(appInfos.Type))
	d.Set("description", appInfos.Description)
	d.Set("version", appInfos.Version)
	d.Set("apptemplate64", appInfos.AppTemplate64)
	d.Set("permissions", permissionsUnixString(*appInfos.Permissions))

	vmTemplate64, _ := appInfos.Template.GetStr("vmtemplate64")
	if err == nil {
		d.Set("vmtemplate64", vmTemplate64)
	}

	publisher, _ := appInfos.Template.GetStr("publisher")
	if err == nil {
		d.Set("publisher", publisher)
	}

	flattenDiags := flattenMarketPlaceAppTemplate(d, meta, &appInfos.Template.Template)
	if len(flattenDiags) > 0 {
		diags = append(diags, flattenDiags...)
		return diags
	}

	state, _ := appInfos.StateString()
	d.Set("disabled", state == app.Disabled.String())

	if appInfos.LockInfos != nil {
		d.Set("lock", LockLevelToString(appInfos.LockInfos.Locked))
	}

	return nil
}

func flattenMarketPlaceAppTemplate(d *schema.ResourceData, meta interface{}, appTpl *dyn.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	err := flattenTemplateSection(d, meta, appTpl)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to read template section",
			Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
		})
	}

	flattenDiags := flattenTemplateTags(d, meta, appTpl)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
	}

	return diags
}

func resourceOpennebulaMarketPlaceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	ac, err := getMarketPlaceAppController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the appliance controller",
			Detail:   err.Error(),
		})
		return diags
	}

	if d.HasChange("") {

	}

	// template management

	appInfos, err := ac.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve appliance informations",
			Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	lock, lockOk := d.GetOk("lock")
	if d.HasChange("lock") && lockOk && lock.(string) == "UNLOCK" {

		err = ac.Unlock()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to unlock",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		ac.Rename(newName)
	}

	update := false
	newTpl := appInfos.Template

	if d.HasChange("origin_id") {
		newTpl.Del(string(appk.OriginID))

		originID := d.Get("origin_id").(int)
		newTpl.AddPair(string(appk.OriginID), originID)

		update = true
	}

	if d.HasChange("type") {
		newTpl.Del(string(appk.Type))

		appType := d.Get("type").(int)
		newTpl.AddPair(string(appk.Type), appType)

		update = true
	}

	if d.HasChange("description") {
		newTpl.Del(string(appk.Description))

		description := d.Get("description").(int)
		newTpl.AddPair(string(appk.Description), description)

		update = true
	}

	if d.HasChange("publisher") {
		newTpl.Del(string(appk.Publisher))

		publisher := d.Get("publisher").(int)
		newTpl.AddPair(string(appk.Publisher), publisher)

		update = true
	}

	if d.HasChange("version") {
		newTpl.Del(string(appk.Version))

		version := d.Get("version").(int)
		newTpl.AddPair(string(appk.Version), version)

		update = true
	}

	if d.HasChange("vmtemplate64") {
		newTpl.Del(string(appk.VMTemplate64))

		vmTemplate := d.Get("vmtemplate64").(int)
		newTpl.AddPair(string(appk.VMTemplate64), vmTemplate)

		update = true
	}

	if d.HasChange("apptemplate64") {
		newTpl.Del(string(appk.AppTemplate64))

		appTemplate := d.Get("apptemplate64").(int)
		newTpl.AddPair(string(appk.AppTemplate64), appTemplate)

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
		err = ac.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update appliance content",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = ac.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change permissions",
					Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
	}

	if d.Get("group") != "" {
		err := changeApplianceGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("marketplace appliance(ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("disabled") {
		disabled := d.Get("disabled").(bool)
		err := ac.Enable(!disabled)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to enable/disable the appliance",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		// wait on state transition
		timeout := d.Timeout(schema.TimeoutUpdate)

		// expected state when disabling
		pendingStates := []string{app.Ready.String()}
		targetStates := []string{app.Disabled.String()}

		// expected states when enabling
		if disabled {
			tmp := pendingStates
			pendingStates = targetStates
			targetStates = tmp
		}

		_, err = waitForMarketAppStates(ctx, ac, timeout, pendingStates, targetStates)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Failed to wait appliance to be in %s state", strings.Join(targetStates, ", ")),
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("lock") && lockOk && lock.(string) != "UNLOCK" {

		var level shared.LockLevel

		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to convert lock level",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = ac.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaMarketPlaceAppRead(ctx, d, meta)
}

func changeApplianceGroup(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller
	var gid int

	ac, err := getMarketPlaceAppController(d, meta)
	if err != nil {
		return err
	}

	group := d.Get("group").(string)
	gid, err = controller.Groups().ByName(group)
	if err != nil {
		return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
	}

	err = ac.Chown(-1, gid)
	if err != nil {
		return fmt.Errorf("Can't find a group with ID `%d`: %s", gid, err)
	}

	return nil
}

func resourceOpennebulaMarketPlaceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	ac, err := getMarketPlaceAppController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the appliance controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = ac.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return nil
}

// waitForMarketAppStates wait for a an market place appliance to reach some expected states
func waitForMarketAppStates(ctx context.Context, hc *goca.MarketPlaceAppController, timeout time.Duration, pending, target []string) (interface{}, error) {

	stateChangeConf := resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
		Refresh: func() (interface{}, string, error) {

			log.Println("Refreshing appliance state...")

			appInfos, err := hc.Info(false)
			if err != nil {
				if NoExists(err) {
					return appInfos, "notfound", nil
				}
				return appInfos, "", err
			}
			state, _ := appInfos.StateString()

			log.Printf("Appliance (ID:%d, name:%s) is currently in state %s", appInfos.ID, appInfos.Name, state)

			// In case we are in some failure state, we try to retrieve more error informations from the appliance template
			if state == app.Error.String() {
				appErr, _ := appInfos.Template.GetStr("ERROR")
				return appInfos, state, fmt.Errorf("Appliance (ID:%d) entered fail state, error: %s", appInfos.ID, appErr)
			}

			return appInfos, state, nil
		},
	}

	return stateChangeConf.WaitForStateContext(ctx)

}
