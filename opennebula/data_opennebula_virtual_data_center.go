package opennebula

import (
	"fmt"
	"strconv"

	vdcSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vdc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaVirtualDataCenter() *schema.Resource {
	return &schema.Resource{
		Read: datasourceOpennebulaVirtualDataCenterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the VDC",
			},
			"tags": tagsSchema(),
		},
	}
}

func vdcFilter(d *schema.ResourceData, meta interface{}) (*vdcSc.VDC, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	vdcs, err := controller.VDCs().Info()
	if err != nil {
		return nil, err
	}

	// filter vdcs with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*vdcSc.VDC, 0, 1)
	for i, vdc := range vdcs.VDCs {

		if nameOk && vdc.Name != name {
			continue
		}

		if tagsOk && !matchTags(vdc.Template.Template, tags) {
			continue
		}

		match = append(match, &vdcs.VDCs[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no virtual data center match the tags")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several virtual data centers match the tags")
	}

	return match[0], nil
}

func datasourceOpennebulaVirtualDataCenterRead(d *schema.ResourceData, meta interface{}) error {

	vdc, err := vdcFilter(d, meta)
	if err != nil {
		return err
	}

	tplPairs := pairsToMap(vdc.Template.Template)

	d.SetId(strconv.FormatInt(int64(vdc.ID), 10))
	d.Set("name", vdc.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			return err
		}
	}

	return nil
}
