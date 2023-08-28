package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"

	app "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/marketplaceapp"
	appk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/marketplaceapp/keys"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

var marketplaceAppType = []string{"IMAGE", "VMTEMPLATE", "SERVICE_TEMPLATE"}

func resourceOpennebulaMarketPlaceApp() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaMarketPlaceAppCreate,
		ReadContext:   resourceOpennebulaMarketPlaceAppRead,
		UpdateContext: resourceOpennebulaMarketPlaceAppUpdate,
		DeleteContext: resourceOpennebulaMarketPlaceAppDelete,
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
			"origin_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The ID of the source image",
				Default:     -1,
			},
			"type": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of Admin user IDs part of the marketplaceApp",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if !contains(value, marketplaceAppType) {
						errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(marketplaceAppType, ",")))
					}

					return
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Text description of the market place appliance",
			},
			"publisher": {
				Type:     schema.TypeString,
				Optional: true,
				//Computed:    true,
				Description: "Publisher of the appliance, if not provided the username will be used",
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
	var ac *goca.MarketPlaceAppController

	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 0)
		if err != nil {
			return nil, err
		}
		ac = controller.MarketPlaceApp(int(gid))
	}

	return ac, nil
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
		tpl.Add(appk.OriginID, val.(string))
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

	tplStr := tpl.String()
	log.Printf("[INFO] Market place appliance definition: %s", tplStr)

	marketID := d.Get("market_id").(int)

	marketplaceAppID, err := controller.MarketPlaceApps().Create(tplStr, marketID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the appliance",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%v", marketplaceAppID))

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

		err = controller.MarketPlaceApp(marketplaceAppID).Lock(level)
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

	marketplaceApp, err := ac.Info(false)
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

	d.Set("name", marketplaceApp.Name)
	d.Set("origin_id", marketplaceApp.OriginID)
	d.Set("type", marketplaceApp.Type)
	d.Set("description", marketplaceApp.Description)
	d.Set("version", marketplaceApp.Version)
	d.Set("apptemplate64", marketplaceApp.AppTemplate64)

	vmTemplate64, _ := marketplaceApp.Template.GetStr("vmtemplate64")
	if err == nil {
		d.Set("apptemplate64", vmTemplate64)
	}

	publisher, _ := marketplaceApp.Template.GetStr("publisher")
	if err == nil {
		d.Set("publisher", publisher)
	}

	flattenDiags := flattenMarketPlaceAppTemplate(d, meta, &marketplaceApp.Template)
	if len(flattenDiags) > 0 {
		diags = append(diags, flattenDiags...)
		return diags
	}

	if marketplaceApp.LockInfos != nil {
		d.Set("lock", LockLevelToString(marketplaceApp.LockInfos.Locked))
	}

	return nil
}

func flattenMarketPlaceAppTemplate(d *schema.ResourceData, meta interface{}, marketplaceAppTpl *dyn.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	err := flattenTemplateSection(d, meta, marketplaceAppTpl)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to read template section",
			Detail:   fmt.Sprintf("marketplace appliance (ID: %s): %s", d.Id(), err),
		})
	}

	flattenDiags := flattenTemplateTags(d, meta, marketplaceAppTpl)
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

	marketplaceApp, err := ac.Info(false)
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
	newTpl := marketplaceApp.Template

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

		updateTemplateSection(d, &newTpl)

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
