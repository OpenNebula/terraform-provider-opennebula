package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vdc"
)

type vdcResources struct {
	ClusterIDs   []int
	HostIDs      []int
	DatastoreIDs []int
	VNetIDs      []int
}

func resourceOpennebulaVirtualDataCenter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualDataCenterCreate,
		ReadContext:   resourceOpennebulaVirtualDataCenterRead,
		UpdateContext: resourceOpennebulaVirtualDataCenterUpdate,
		DeleteContext: resourceOpennebulaVirtualDataCenterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC",
			},
			"group_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "List of Group IDs to be added into the VDC",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"zones": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of zones to add into the VDC",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Resources Zone ID (default: 0)",
						},
						"host_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Computed:    true,
							Description: "List of Host IDs from the Zone to add in the VDC",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"datastore_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Computed:    true,
							Description: "List of Datastore IDs from the Zone to add in the VDC",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"vnet_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Computed:    true,
							Description: "List of VNET IDs from the Zone to add in the VDC",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"cluster_ids": {
							Type:        schema.TypeSet,
							Optional:    true,
							Computed:    true,
							Description: "List of cluster IDs from the Zone to add in the VDC",
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
					},
				},
			},
		},
	}
}

func getVDCController(d *schema.ResourceData, meta interface{}) (*goca.VDCController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	vdcID, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.VDC(int(vdcID)), nil
}

func resourceOpennebulaVirtualDataCenterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	vdcDef, err := generateVDC(d)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate description",
			Detail:   err.Error(),
		})
		return diags
	}

	vdcID, err := controller.VDCs().Create(vdcDef, -1)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the virtual data center",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%v", vdcID))

	vdcc := controller.VDC(vdcID)

	zones := d.Get("zones").(*schema.Set).List()
	for i := 0; i < len(zones); i++ {
		zMap := zones[i].(map[string]interface{})
		zone_id := zMap["id"].(int)
		hosts := zMap["host_ids"].(*schema.Set).List()
		datastores := zMap["datastore_ids"].(*schema.Set).List()
		clusters := zMap["cluster_ids"].(*schema.Set).List()
		vnets := zMap["vnet_ids"].(*schema.Set).List()

		// Add Hosts from the zone
		for j := 0; j < len(hosts); j++ {
			err = vdcc.AddHost(zone_id, hosts[j].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add host",
					Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		// Add Datastore from the zone
		for j := 0; j < len(datastores); j++ {
			err = vdcc.AddDatastore(zone_id, datastores[j].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add datastore",
					Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		// Add clusters from the zone
		for j := 0; j < len(clusters); j++ {
			err = vdcc.AddCluster(zone_id, clusters[j].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add cluster",
					Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		// Add vnets from the zone
		for j := 0; j < len(vnets); j++ {
			err = vdcc.AddVnet(zone_id, vnets[j].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add virtual network",
					Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
	}

	// add groups if list provided
	if groupids, ok := d.GetOk("group_ids"); ok {
		grouplist := groupids.(*schema.Set).List()
		for i := 0; i < len(grouplist); i++ {
			err = vdcc.AddGroup(grouplist[i].(int))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add group",
					Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
	}

	return resourceOpennebulaVirtualDataCenterRead(ctx, d, meta)
}

func resourceOpennebulaVirtualDataCenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vdcc, err := getVDCController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the virtual data center controller",
			Detail:   err.Error(),
		})
		return diags

	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vdc, err := vdcc.Info(false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual data center %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", vdc.ID))
	d.Set("name", vdc.Name)
	d.Set("zones", flattenZones(vdc))
	d.Set("group_ids", vdc.Groups.ID)

	return nil
}

func getAddDelIntList(ngrouplist, ogrouplist []interface{}) ([]int, []int) {
	addgroup := []int{}
	// Get new groups to add
	for _, ngroup := range ngrouplist {
		found := false
		for _, ogroup := range ogrouplist {
			if ngroup.(int) == ogroup.(int) {
				found = true
				break
			}
			if !found {
				addgroup = append(addgroup, ngroup.(int))
			}
		}
	}
	// Get old groups to delete
	delgroup := []int{}
	for _, ogroup := range ogrouplist {
		found := false
		for _, ngroup := range ngrouplist {
			if ogroup.(int) == ngroup.(int) {
				found = true
				break
			}
			if !found {
				delgroup = append(delgroup, ogroup.(int))
			}
		}
	}

	return addgroup, delgroup
}

func resourceOpennebulaVirtualDataCenterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vdcc, err := getVDCController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the virtual data center controller",
			Detail:   err.Error(),
		})
		return diags
	}

	if d.HasChange("name") {
		// Rename VDC
		err = vdcc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("group_ids") {
		ogroups, ngroups := d.GetChange("group_ids")
		ogrouplist := ogroups.(*schema.Set).List()
		ngrouplist := ngroups.(*schema.Set).List()

		addgroup, delgroup := getAddDelIntList(ngrouplist, ogrouplist)

		// Delete old groups
		for _, g := range delgroup {
			err = vdcc.DelGroup(g)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to delete group",
					Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		// Add new groups
		for _, g := range addgroup {
			err = vdcc.AddGroup(g)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add group",
					Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
	}

	if d.HasChange("zones") {
		ozonesset, nzonesset := d.GetChange("zones")
		ozones := ozonesset.(*schema.Set).List()
		nzones := nzonesset.(*schema.Set).List()

		// Delete all old zones
		for _, ozone := range ozones {
			ozMap := ozone.(map[string]interface{})
			// This is an old zone id to delete
			zone_id := ozMap["id"].(int)
			hosts := ozMap["host_ids"].(*schema.Set).List()
			datastores := ozMap["datastore_ids"].(*schema.Set).List()
			clusters := ozMap["cluster_ids"].(*schema.Set).List()
			vnets := ozMap["vnet_ids"].(*schema.Set).List()

			// Delete Hosts for the zone
			for j := 0; j < len(hosts); j++ {
				err = vdcc.DelHost(zone_id, hosts[j].(int))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to delete host",
						Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
			// Delele Datastore from the zone
			for j := 0; j < len(datastores); j++ {
				err = vdcc.DelDatastore(zone_id, datastores[j].(int))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to delete datastore",
						Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
			// Delete clusters from the zone
			for j := 0; j < len(clusters); j++ {
				err = vdcc.DelCluster(zone_id, clusters[j].(int))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to delete cluster",
						Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
			// Delete vnets from the zone
			for j := 0; j < len(vnets); j++ {
				err = vdcc.DelVnet(zone_id, vnets[j].(int))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to delete virtual network",
						Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
		}

		// Get Add new zone
		for _, nzone := range nzones {
			nzMap := nzone.(map[string]interface{})
			// This is a new zone id
			zone_id := nzMap["id"].(int)
			hosts := nzMap["host_ids"].(*schema.Set).List()
			datastores := nzMap["datastore_ids"].(*schema.Set).List()
			clusters := nzMap["cluster_ids"].(*schema.Set).List()
			vnets := nzMap["vnet_ids"].(*schema.Set).List()

			// Add Hosts from the zone
			for j := 0; j < len(hosts); j++ {
				err = vdcc.AddHost(zone_id, hosts[j].(int))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to add host",
						Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
			// Add Datastore from the zone
			for j := 0; j < len(datastores); j++ {
				err = vdcc.AddDatastore(zone_id, datastores[j].(int))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to add datastore",
						Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
			// Add clusters from the zone
			for j := 0; j < len(clusters); j++ {
				err = vdcc.AddCluster(zone_id, clusters[j].(int))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to add cluster",
						Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
			// Add vnets from the zone
			for j := 0; j < len(vnets); j++ {
				err = vdcc.AddVnet(zone_id, vnets[j].(int))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to add virtual network",
						Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
		}
	}

	return resourceOpennebulaVirtualDataCenterRead(ctx, d, meta)
}

func flattenZones(vdc *vdc.VDC) []map[string]interface{} {

	zones := make(map[int]*vdcResources, 0)

	// Get clusters
	for _, cluster := range vdc.Clusters {
		if zonecluster, ok := zones[cluster.ZoneID]; ok {
			zonecluster.ClusterIDs = append(zonecluster.ClusterIDs, cluster.ClusterID)
		} else {
			zones[cluster.ZoneID] = &vdcResources{
				ClusterIDs: []int{cluster.ClusterID},
			}
		}
	}
	// Get hosts
	for _, host := range vdc.Hosts {
		if zonehost, ok := zones[host.ZoneID]; ok {
			zonehost.HostIDs = append(zonehost.HostIDs, host.HostID)
		} else {
			zones[host.ZoneID] = &vdcResources{
				HostIDs: []int{host.HostID},
			}
		}
	}
	// Get datastores
	for _, ds := range vdc.Datastores {
		if zoneds, ok := zones[ds.ZoneID]; ok {
			zoneds.DatastoreIDs = append(zoneds.DatastoreIDs, ds.DatastoreID)
		} else {
			zones[ds.ZoneID] = &vdcResources{
				DatastoreIDs: []int{ds.DatastoreID},
			}
		}
	}
	// Get vnet
	for _, vnet := range vdc.VNets {
		if zonevnet, ok := zones[vnet.ZoneID]; ok {
			zonevnet.VNetIDs = append(zonevnet.VNetIDs, vnet.VnetID)
		} else {
			zones[vnet.ZoneID] = &vdcResources{
				VNetIDs: []int{vnet.VnetID},
			}
		}
	}

	zonemap := make([]map[string]interface{}, 0)
	for k, v := range zones {
		zmap := map[string]interface{}{
			"id":            k,
			"cluster_ids":   v.ClusterIDs,
			"host_ids":      v.HostIDs,
			"datastore_ids": v.DatastoreIDs,
			"vnet_ids":      v.VNetIDs,
		}
		zonemap = append(zonemap, zmap)
	}

	return zonemap
}

func resourceOpennebulaVirtualDataCenterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vdcc, err := getVDCController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the virtual data center controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = vdcc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("virtual data center (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return nil
}

func generateVDC(d *schema.ResourceData) (string, error) {

	tpl := dyn.NewTemplate()

	name := d.Get("name").(string)
	tpl.AddPair("NAME", name)

	tplStr := tpl.String()
	log.Printf("[INFO] VDC definition: %s", tplStr)

	return tplStr, nil
}
