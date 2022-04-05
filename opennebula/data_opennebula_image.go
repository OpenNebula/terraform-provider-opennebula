package opennebula

import (
	"fmt"
	"strconv"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	imageSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataOpennebulaImage() *schema.Resource {
	return &schema.Resource{
		Read: datasourceOpennebulaImageRead,

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

	controller := meta.(*goca.Controller)

	images, err := controller.Images().Info()
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
		return nil, fmt.Errorf("no image match the tags")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several images match the tags")
	}

	return match[0], nil
}

func datasourceOpennebulaImageRead(d *schema.ResourceData, meta interface{}) error {

	image, err := imageFilter(d, meta)
	if err != nil {
		return err
	}

	tplPairs := pairsToMap(image.Template.Template)

	d.SetId(strconv.FormatInt(int64(image.ID), 10))
	d.Set("name", image.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			return err
		}
	}

	return nil
}
