package opennebula

import (
	"context"
	"fmt"
	"strconv"

	marketplaceSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/marketplace"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaMarketplace() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaMarketplaceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "Id of the marketplace",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the marketplace",
			},
			"tags": tagsSchema(),
		},
	}
}

func marketplaceFilter(d *schema.ResourceData, meta interface{}) (*marketplaceSc.MarketPlace, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	marketplaces, err := controller.MarketPlaces().Info()
	if err != nil {
		return nil, err
	}

	// filter marketplaces with user defined criterias
	id := d.Get("id")
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*marketplaceSc.MarketPlace, 0, 1)
	for i, marketplace := range marketplaces.MarketPlaces {

		if id != -1 && marketplace.ID != id {
			continue
		}

		if nameOk && marketplace.Name != name {
			continue
		}

		if tagsOk && !matchTags(marketplace.Template.Template, tags) {
			continue
		}

		match = append(match, &marketplaces.MarketPlaces[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no marketplace match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several marketplaces match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaMarketplaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	marketplace, err := marketplaceFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "marketplaces filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(marketplace.Template.Template)

	d.SetId(strconv.FormatInt(int64(marketplace.ID), 10))
	d.Set("name", marketplace.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("Marketplace (ID: %d): %s", marketplace.ID, err),
			})
			return diags
		}
	}

	return nil
}
