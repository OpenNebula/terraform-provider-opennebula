package opennebula

import (
	"context"
	"fmt"
	"strconv"

	datastoreSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/datastore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaDatastore() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaDatastoreRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the datastore",
			},
			"tags": tagsSchema(),
		},
	}
}

func datastoreFilter(d *schema.ResourceData, meta interface{}) (*datastoreSc.Datastore, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	datastores, err := controller.Datastores().Info()
	if err != nil {
		return nil, err
	}

	// filter datastores with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*datastoreSc.Datastore, 0, 1)
	for i, datastore := range datastores.Datastores {

		if nameOk && datastore.Name != name {
			continue
		}

		if tagsOk && !matchTags(datastore.Template.Template, tags) {
			continue
		}

		match = append(match, &datastores.Datastores[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no datastore match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several datastores match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaDatastoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	datastore, err := datastoreFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "datastores filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(datastore.Template.Template)

	d.SetId(strconv.FormatInt(int64(datastore.ID), 10))
	d.Set("name", datastore.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("Datastore (ID: %d): %s", datastore.ID, err),
			})
			return diags
		}
	}

	return nil
}
