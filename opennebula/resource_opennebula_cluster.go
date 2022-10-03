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
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/cluster"
)

func resourceOpennebulaCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaClusterCreate,
		ReadContext:   resourceOpennebulaClusterRead,
		UpdateContext: resourceOpennebulaClusterUpdate,
		DeleteContext: resourceOpennebulaClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Cluster",
			},
			"hosts": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of hosts IDs part of the cluster",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"datastores": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of datastores IDs part of the cluster",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"virtual_networks": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of virtual network IDs part of the cluster",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tags":         tagsSchema(),
			"default_tags": defaultTagsSchemaComputed(),
			"tags_all":     tagsSchemaComputed(),
		},
	}
}

func getClusterController(d *schema.ResourceData, meta interface{}) (*goca.ClusterController, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	var gc *goca.ClusterController

	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 0)
		if err != nil {
			return nil, err
		}
		gc = controller.Cluster(int(gid))
	}

	return gc, nil
}

func resourceOpennebulaClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	clusterID, err := controller.Clusters().Create(d.Get("name").(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the cluster",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%d", clusterID))

	cc := controller.Cluster(clusterID)

	// add hosts
	if hostsIf, ok := d.GetOk("hosts"); ok {
		hostsList := hostsIf.(*schema.Set).List()
		for i := 0; i < len(hostsList); i++ {
			err = cc.AddHost(hostsList[i].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add hosts",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", clusterID, err),
				})
				return diags
			}
		}
	}

	// add datastores
	if datastoreIf, ok := d.GetOk("datastores"); ok {
		datastoreList := datastoreIf.(*schema.Set).List()
		for i := 0; i < len(datastoreList); i++ {
			err = cc.AddDatastore(datastoreList[i].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add datastore",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", clusterID, err),
				})
				return diags
			}
		}
	}

	// add virtual networks
	if vnetIf, ok := d.GetOk("virtual_networks"); ok {
		vnetList := vnetIf.(*schema.Set).List()
		for i := 0; i < len(vnetList); i++ {
			err = cc.AddVnet(vnetList[i].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add virtual network",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", clusterID, err),
				})
				return diags
			}
		}
	}

	// template management

	tpl := dyn.NewTemplate()

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

	if len(tpl.Elements) > 0 {
		err = cc.Update(tpl.String(), parameters.Merge)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update the cluster content",
				Detail:   fmt.Sprintf("cluster (ID: %d): %s", clusterID, err),
			})
			return diags
		}
	}

	return resourceOpennebulaClusterRead(ctx, d, meta)
}

func resourceOpennebulaClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	cc, err := getClusterController(d, meta)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing cluster %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the cluster controller",
			Detail:   err.Error(),
		})
		return diags
	}

	clusterInfos, err := cc.Info()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed retrieve cluster informations",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.Set("name", clusterInfos.Name)

	// read cluster hosts members
	hostsIDs := make([]int, 0)
	for _, id := range clusterInfos.Hosts.ID {
		hostsIDs = append(hostsIDs, id)
	}

	err = d.Set("hosts", hostsIDs)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set hosts field",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// read cluster datastore members
	datastoreIDs := make([]int, 0)
	for _, id := range clusterInfos.Datastores.ID {
		datastoreIDs = append(datastoreIDs, id)
	}

	err = d.Set("datastores", datastoreIDs)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set datastores field",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// read cluster virtual network members
	vnetIDs := make([]int, 0)
	for _, id := range clusterInfos.Vnets.ID {
		vnetIDs = append(vnetIDs, id)
	}

	err = d.Set("virtual_networks", vnetIDs)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set virtual_networks field",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	err = flattenClusterTemplate(d, meta, &clusterInfos.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten template",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return nil
}

func flattenClusterTemplate(d *schema.ResourceData, meta interface{}, clusterTpl *cluster.Template) error {
	config := meta.(*Configuration)

	tags := make(map[string]interface{})
	tagsAll := make(map[string]interface{})

	// Get default tags
	oldDefault := d.Get("default_tags").(map[string]interface{})
	for k, _ := range oldDefault {
		tagValue, err := clusterTpl.GetStr(strings.ToUpper(k))
		if err != nil {
			return nil
		}
		tagsAll[k] = tagValue
	}
	d.Set("default_tags", config.defaultTags)

	// Get only tags described in the configuration
	if tagsInterface, ok := d.GetOk("tags"); ok {

		for k, _ := range tagsInterface.(map[string]interface{}) {
			tagValue, err := clusterTpl.GetStr(strings.ToUpper(k))
			if err != nil {
				return err
			}
			tags[k] = tagValue
			tagsAll[k] = tagValue
		}

		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
	}
	d.Set("tags_all", tagsAll)

	return nil
}

func resourceOpennebulaClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	cc, err := getClusterController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the cluster controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// template management

	cluster, err := cc.Info()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve clusters informations",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	update := false
	newTpl := cluster.Template

	if d.HasChange("hosts") {

		oldHostsIf, newHostsIf := d.GetChange("hosts")

		oldHosts := schema.NewSet(schema.HashInt, oldHostsIf.(*schema.Set).List())
		newHosts := schema.NewSet(schema.HashInt, newHostsIf.(*schema.Set).List())

		// delete hosts
		remHosts := oldHosts.Difference(newHosts)

		for _, id := range remHosts.List() {

			// we need to check is the ID is not an old ID
			// i.e. the ID of an user deleted/replaced
			_, err := controller.Host(id.(int)).Info(false)
			if err != nil {
				if NoExists(err) {
					continue
				}
			}

			err = cc.DelHost(id.(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to delete a host from the cluster",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", cc.ID, err),
				})
				return diags
			}
		}

		// add hosts
		addHosts := newHosts.Difference(oldHosts)

		for _, id := range addHosts.List() {
			err := cc.AddHost(id.(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add a host to the cluster",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", cc.ID, err),
				})
				return diags
			}
		}
	}

	if d.HasChange("datastores") {

		oldDatastoresIf, newDatastoresIf := d.GetChange("datastores")

		oldDatastores := schema.NewSet(schema.HashInt, oldDatastoresIf.(*schema.Set).List())
		newDatastores := schema.NewSet(schema.HashInt, newDatastoresIf.(*schema.Set).List())

		// delete datastores
		remDatastores := oldDatastores.Difference(newDatastores)

		for _, id := range remDatastores.List() {

			// we need to check is the ID is not an old ID
			// i.e. the ID of an user deleted/replaced
			_, err := controller.Datastore(id.(int)).Info(false)
			if err != nil {
				if NoExists(err) {
					continue
				}
			}

			err = cc.DelDatastore(id.(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to delete a datastore from the cluster",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", cc.ID, err),
				})
				return diags
			}
		}

		// add datastores
		addDatastores := newDatastores.Difference(oldDatastores)

		for _, id := range addDatastores.List() {
			err := cc.AddDatastore(id.(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add a datastore to the cluster",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", cc.ID, err),
				})
				return diags
			}
		}
	}

	if d.HasChange("virtual_networks") {

		oldVNetIf, newVNetIf := d.GetChange("virtual_networks")

		oldVNet := schema.NewSet(schema.HashInt, oldVNetIf.(*schema.Set).List())
		newVNet := schema.NewSet(schema.HashInt, newVNetIf.(*schema.Set).List())

		// delete virtual network
		remVNet := oldVNet.Difference(newVNet)

		for _, id := range remVNet.List() {

			// we need to check is the ID is not an old ID
			// i.e. the ID of an user deleted/replaced
			_, err := controller.VirtualNetwork(id.(int)).Info(false)
			if err != nil {
				if NoExists(err) {
					continue
				}
			}

			// delete the user from group admin
			err = cc.DelVnet(id.(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to delete a virtual network from the cluster",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", cc.ID, err),
				})
				return diags
			}
		}

		// add virtual networks
		addVNet := newVNet.Difference(oldVNet)

		for _, id := range addVNet.List() {
			err := cc.AddVnet(id.(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add a virtual network to the cluster",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", cc.ID, err),
				})
				return diags
			}
		}
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
		err = cc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update cluster content",
				Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

	}

	return resourceOpennebulaClusterRead(ctx, d, meta)
}

func resourceOpennebulaClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	cc, err := getClusterController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the cluster controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = cc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("cluster (ID: %d): %s", cc.ID, err),
		})
		return diags
	}

	return nil
}
