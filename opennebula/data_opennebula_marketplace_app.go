package opennebula

import (
	"context"
	"fmt"
	"strconv"

	appSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/marketplaceapp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaMarketplaceApp() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaMarketplaceAppRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Id of the appliance",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the appliance",
			},
			"tags": tagsSchema(),
		},
	}
}

func applianceFilter(d *schema.ResourceData, meta interface{}) (*appSc.MarketPlaceApp, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	apps, err := controller.MarketPlaceApps().Info()
	if err != nil {
		return nil, err
	}

	// filter appliances
	id := d.Get("id")
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*appSc.MarketPlaceApp, 0, 1)
	for i, app := range apps.MarketPlaceApps {

		if id != -1 && app.ID != id {
			continue
		}

		if nameOk && app.Name != name {
			continue
		}

		if tagsOk && !matchTags(app.Template.Template, tags) {
			continue
		}

		match = append(match, &apps.MarketPlaceApps[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no appliance match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several appliances match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaMarketplaceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	app, err := applianceFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "appliances filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(app.Template.Template)

	d.SetId(strconv.FormatInt(int64(app.ID), 10))
	d.Set("name", app.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("Appliance (ID: %d): %s", app.ID, err),
			})
			return diags
		}
	}

	return nil
}
