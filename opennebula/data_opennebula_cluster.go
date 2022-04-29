package opennebula

import (
	"fmt"
	"strconv"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	clusterSc "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/cluster"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaCluster() *schema.Resource {
	return &schema.Resource{
		Read: datasourceOpennebulaClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the cluster",
			},
			"tags": tagsSchema(),
		},
	}
}

func clusterFilter(d *schema.ResourceData, meta interface{}) (*clusterSc.Cluster, error) {

	controller := meta.(*goca.Controller)

	clusters, err := controller.Clusters().Info(false)
	if err != nil {
		return nil, err
	}

	// filter clusters with user defined criterias
	name, nameOk := d.GetOk("name")
	tagsInterface, tagsOk := d.GetOk("tags")
	tags := tagsInterface.(map[string]interface{})

	match := make([]*clusterSc.Cluster, 0, 1)
	for i, cluster := range clusters.Clusters {

		if nameOk && cluster.Name != name {
			continue
		}

		if tagsOk && !matchTags(cluster.Template.Template, tags) {
			continue
		}

		match = append(match, &clusters.Clusters[i])
	}

	// check filtering results
	if len(match) == 0 {
		return nil, fmt.Errorf("no cluster match the tags")
	} else if len(match) > 1 {
		return nil, fmt.Errorf("several clusters match the tags")
	}

	return match[0], nil
}

func datasourceOpennebulaClusterRead(d *schema.ResourceData, meta interface{}) error {

	cluster, err := clusterFilter(d, meta)
	if err != nil {
		return err
	}

	tplPairs := pairsToMap(cluster.Template.Template)

	d.SetId(strconv.FormatInt(int64(cluster.ID), 10))
	d.Set("name", cluster.Name)

	if len(tplPairs) > 0 {
		err := d.Set("tags", tplPairs)
		if err != nil {
			return err
		}
	}

	return nil
}
