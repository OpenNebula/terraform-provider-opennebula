package opennebula

import (
	"fmt"
	"strconv"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	vmGroupSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vmgroup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataOpennebulaVMGroup() *schema.Resource {
	return &schema.Resource{
		Read: datasourceOpennebulaVMGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Virtual Machine Group",
			},
			"tags": tagsSchema(),
		},
	}
}

func vmGroupFilter(d *schema.ResourceData, meta interface{}) (*vmGroupSc.VMGroup, error) {

	controller := meta.(*goca.Controller)

	vmGroups, err := controller.VMGroups().Info()
	if err != nil {
		return nil, err
	}

	// filter vm groups with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*vmGroupSc.VMGroup, 0, 1)
	for i, vmGroup := range vmGroups.VMGroups {

		if nameOk && vmGroup.Name != name {
			continue
		}

		if tagsOk && !matchTags(vmGroup.Template, tags) {
			continue
		}

		match = append(match, &vmGroups.VMGroups[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no vm group match the tags")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several vm group match the tags")
	}

	return match[0], nil
}

func datasourceOpennebulaVMGroupRead(d *schema.ResourceData, meta interface{}) error {

	vmGroup, err := vmGroupFilter(d, meta)
	if err != nil {
		return err
	}

	tplPairs := pairsToMap(vmGroup.Template)

	d.SetId(strconv.FormatInt(int64(vmGroup.ID), 10))
	d.Set("name", vmGroup.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			return err
		}
	}

	return nil
}
