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
				Optional:    false,
				Computed:    true,
				Description: "List of hosts IDs part of the cluster",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"datastores": {
				Type:        schema.TypeSet,
				Optional:    false,
				Computed:    true,
				Description: "List of datastores IDs part of the cluster",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Deprecated: "use cluster_ids field from the datastore resource instead",
			},
			"virtual_networks": {
				Type:        schema.TypeSet,
				Optional:    false,
				Computed:    true,
				Description: "List of virtual network IDs part of the cluster",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Deprecated: "use cluster_ids field from the virtual network resource instead",
			},
			"tags":             tagsSchema(),
			"default_tags":     defaultTagsSchemaComputed(),
			"tags_all":         tagsSchemaComputed(),
			"template_section": templateSectionSchema(),
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

	// template management

	tpl := dyn.NewTemplate()

	vectorsInterface := d.Get("template_section").(*schema.Set).List()
	if len(vectorsInterface) > 0 {
		addTemplateVectors(vectorsInterface, tpl)
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
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the cluster controller",
			Detail:   err.Error(),
		})
		return diags
	}

	clusterInfos, err := cc.Info()
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing cluster %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed retrieve cluster informations",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.Set("name", clusterInfos.Name)

	// read cluster hosts members
	err = d.Set("hosts", clusterInfos.Hosts.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set hosts field",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// read cluster datastore members
	err = d.Set("datastores", clusterInfos.Datastores.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set datastores field",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// read cluster virtual network members
	err = d.Set("virtual_networks", clusterInfos.Vnets.ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set virtual_networks field",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	flattenDiags := flattenClusterTemplate(d, meta, &clusterInfos.Template)
	for _, diag := range flattenDiags {
		diags = append(diags, diag)
	}

	return diags
}

func flattenClusterTemplate(d *schema.ResourceData, meta interface{}, clusterTpl *cluster.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	err := flattenTemplateSection(d, meta, &clusterTpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to flatten template section",
			Detail:   fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err),
		})
	}

	flattenDiags := flattenTemplateTags(d, meta, &clusterTpl.Template)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("cluster (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
	}

	return diags
}

func resourceOpennebulaClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

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

	if d.HasChange("template_section") {

		updateTemplateSection(d, &newTpl.Template)

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
