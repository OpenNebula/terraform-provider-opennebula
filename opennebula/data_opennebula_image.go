package opennebula

import (
	"context"
	"fmt"
	"strconv"

	imageSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaImage() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaImageRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the image",
			},
			"tags": tagsSchema(),
		},
	}
}

func imageFilter(d *schema.ResourceData, meta interface{}) (*imageSc.Image, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	// search for any images this user can see
	images, err := controller.Images().Info(-2, -1, -1)
	if err != nil {
		return nil, err
	}

	// filter images with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*imageSc.Image, 0, 1)
	for i, image := range images.Images {

		if nameOk && image.Name != name {
			continue
		}

		if tagsOk && !matchTags(image.Template.Template, tags) {
			continue
		}

		match = append(match, &images.Images[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no image match the constraints")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several images match the constraints")
	}

	return match[0], nil
}

func datasourceOpennebulaImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	image, err := imageFilter(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "images filtering failed",
			Detail:   err.Error(),
		})
		return diags
	}

	tplPairs := pairsToMap(image.Template.Template)

	d.SetId(strconv.FormatInt(int64(image.ID), 10))
	d.Set("name", image.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "setting attribute failed",
				Detail:   fmt.Sprintf("Image (ID: %d): %s", image.ID, err),
			})
			return diags
		}
	}

	return nil
}
