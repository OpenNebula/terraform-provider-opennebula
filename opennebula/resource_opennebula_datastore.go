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
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/datastore"
	dsKey "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/datastore/keys"
)

var datastoreTypes = map[string]string{
	"IMAGE":  "IMAGE_DS",
	"SYSTEM": "SYSTEM_DS",
	"FILE":   "FILE_DS"}

func resourceOpennebulaDatastore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaDatastoreCreate,
		ReadContext:   resourceOpennebulaDatastoreRead,
		UpdateContext: resourceOpennebulaDatastoreUpdate,
		DeleteContext: resourceOpennebulaDatastoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Datastore",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Type of the datastore: image, system, files",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := strings.ToUpper(v.(string))

					keys := make([]string, 0, len(datastoreTypes))
					for k := range datastoreTypes {
						keys = append(keys, k)
					}
					if !contains(value, keys) {
						errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(keys, ",")))
					}

					return
				},
			},
			"cluster_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of cluster IDs hosting the datastore, if not set it uses the default cluster",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				MinItems: 1,
			},
			"restricted_directories": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Paths that cannot be used to register images. A space separated list of paths",
			},
			"safe_directories": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "If you need to allow a directory listed under RESTRICTED_DIRS. A space separated list of paths",
			},
			"no_decompress": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Do not try to untar or decompress the file to be registered",
			},
			"storage_usage_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum capacity allowed for the Datastore in MB",
			},
			"transfer_bandwith_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Specify the maximum transfer rate in bytes/second when downloading images from a http/https URL. Suffixes K, M or G can be used",
			},
			"check_available_capacity": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If yes, the available capacity of the Datastore is checked before creating a new image",
			},
			"bridge_list": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of hosts that have access to the storage to add new images to the datastore",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"staging_dir": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path in the storage bridge host to copy an Image before moving it to its final destination",
			},
			"driver": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Specific image mapping driver enforcement. If present it overrides image DRIVER set in the image attributes and VM template",
			},
			"compatible_system_datastore": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "For Image Datastores only. Set the System Datastores IDs that can be used with an Image Datastore",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"ceph": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pool_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Ceph pool name",
						},
						"user": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Ceph user name",
						},
						"key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Key file for user. if not set default locations are used",
						},
						"config": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Non-default Ceph configuration file if needed",
						},
						"rbd_format": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "By default RBD Format 2 will be used",
						},
						"secret": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The UUID of the libvirt secret",
						},
						"host": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "List of Ceph monitors",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"local_storage": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Use local host storage, SSH mode",
						},
						"trash": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enables trash feature on given datastore",
						},
					},
				},
				ConflictsWith: []string{"custom"},
			},
			"custom": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datastore": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Datastore driver",
						},
						"transfer": {
							Type:        schema.TypeString,
							ForceNew:    true,
							Required:    true,
							Description: "Transfer driver",
						},
					},
				},
				ConflictsWith: []string{"ceph"},
			},
			"tags":         tagsSchema(),
			"default_tags": defaultTagsSchemaComputed(),
			"tags_all":     tagsSchemaComputed(),
		},
	}
}

func getDatastoreController(d *schema.ResourceData, meta interface{}) (*goca.DatastoreController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	dsID, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.Datastore(int(dsID)), nil
}

func resourceOpennebulaDatastoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	tpl := datastore.NewTemplate()

	tpl.Add(dsKey.Name, d.Get("name").(string))
	dsType := strings.ToUpper(d.Get("type").(string))
	tpl.Add(dsKey.Type, datastoreTypes[dsType])

	restrictedDirs := d.Get("restricted_directories").(string)
	if len(restrictedDirs) > 0 {
		tpl.Add(dsKey.RestrictedDirs, restrictedDirs)
	}
	safeDirs := d.Get("safe_directories").(string)
	if len(safeDirs) > 0 {
		tpl.Add(dsKey.SafeDirs, safeDirs)
	}
	storageUsageLimit, ok := d.GetOk("storage_usage_limit")
	if ok {
		tpl.Add(dsKey.LimitMB, storageUsageLimit.(int))
	}
	transferBandwithLimit, ok := d.GetOk("transfer_bandwith_limit")
	if ok {
		tpl.Add(dsKey.LimitTransferBW, transferBandwithLimit.(int))
	}
	noDecompress := d.Get("no_decompress").(bool)
	if noDecompress {
		tpl.Add(dsKey.NoDecompress, noDecompress)
	}
	checkAvailableCapacity := d.Get("check_available_capacity").(bool)
	if checkAvailableCapacity {
		tpl.Add(dsKey.DatastoreCapacityCheck, checkAvailableCapacity)
	}

	bridges := d.Get("bridge_list").(*schema.Set).List()
	if len(bridges) > 0 {

		// convert to slice of strings then join
		var bridgeStrs []string
		for _, bridge := range bridges {
			bridgeStrs = append(bridgeStrs, bridge.(string))
		}

		tpl.Add(dsKey.BridgeList, strings.Join(bridgeStrs, " "))
	}

	stagingDir := d.Get("staging_dir").(string)
	if len(stagingDir) > 0 {
		tpl.Add(dsKey.StagingDir, stagingDir)
	}
	driver := d.Get("driver").(string)
	if len(driver) > 0 {
		tpl.Add(dsKey.Driver, driver)
	}

	compatibleSystemDatastores := d.Get("compatible_system_datastore").(*schema.Set).List()
	if len(compatibleSystemDatastores) > 0 {

		// convert to slice of strings then join
		var compatibleSysDs []string
		for _, sysDs := range compatibleSystemDatastores {
			compatibleSysDs = append(compatibleSysDs, fmt.Sprint(sysDs.(int)))
		}

		tpl.Add(dsKey.CompatibleSysDs, strings.Join(compatibleSysDs, ","))
	}

	customAttrsList := d.Get("custom").(*schema.Set).List()

	// Ceph: https://docs.opennebula.io/6.4/vmware_cluster_deployment/vmware_storage_setup/vcenter_datastores.html#vcenter-ds
	cephAttrsList := d.Get("ceph").(*schema.Set).List()

	if len(cephAttrsList) > 0 {
		cephAttrsMap := cephAttrsList[0].(map[string]interface{})

		if dsType == "IMAGE" {
			tpl.Add("TM_MAD", "ceph")
		} else if dsType == "SYSTEM" {
			if cephAttrsMap["local_storage"].(bool) {
				tpl.Add("TM_MAD", "ssh")
			} else {
				tpl.Add("TM_MAD", "ceph")
			}
		}

		addCephAttributes(cephAttrsMap, tpl)

	} else if len(customAttrsList) > 0 {
		customAttrsMap := customAttrsList[0].(map[string]interface{})
		datastoreDriver, _ := customAttrsMap["datastore"]
		if len(datastoreDriver.(string)) > 0 {
			tpl.Add("DS_MAD", datastoreDriver)
		}
		transferDriver, ok := customAttrsMap["transfer"]
		if ok {
			tpl.Add("TM_MAD", transferDriver)
		}
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

	clusterIDs := d.Get("cluster_ids").(*schema.Set).List()
	if len(clusterIDs) == 0 {
		clusterIDs = []interface{}{-1}
	}

	log.Printf("[INFO] Datastore template: %s\n", tpl.String())
	datastoreID, err := controller.Datastores().Create(tpl.String(), clusterIDs[0].(int))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the datastore",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%v", datastoreID))

	// Set Clusters (first in list is already set)
	if len(clusterIDs) > 1 {
		for _, id := range clusterIDs[1:] {
			err := controller.Cluster(id.(int)).AddDatastore(datastoreID)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to set cluster",
					Detail:   fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

		}
	}

	return resourceOpennebulaDatastoreRead(ctx, d, meta)
}

func addCephAttributes(attrs map[string]interface{}, tpl *datastore.Template) {

	tpl.Add("DS_MAD", "ceph")
	tpl.Add("DISK_TYPE", "RBD")

	poolName := attrs["pool_name"].(string)
	if len(poolName) > 0 {
		tpl.Add("POOL_NAME", poolName)
	}
	user, ok := attrs["user"]
	if ok {
		tpl.Add("CEPH_USER", user)
	}
	key, ok := attrs["key"]
	if ok {
		tpl.Add("CEPH_KEY", key)
	}
	conf, ok := attrs["config"]
	if ok {
		tpl.Add("CEPH_CONF", conf)
	}
	rbdFormat, ok := attrs["rbd_format"]
	if ok {
		tpl.Add("RBD_FORMAT", rbdFormat)
	}
	cephSecret, ok := attrs["secret"]
	if ok {
		tpl.Add("CEPH_SECRET", cephSecret)
	}
	trash, ok := attrs["trash"]
	if ok {
		tpl.Add("CEPH_TRASH", trash)
	}

	hostsIf, ok := attrs["host"]
	if ok {
		hosts := hostsIf.(*schema.Set).List()

		var hostsStr strings.Builder
		for _, host := range hosts {
			hostsStr.WriteString(host.(string))
			hostsStr.WriteString(" ")
		}
		tpl.Add("CEPH_HOST", hostsStr.String())
	}
}

func resourceOpennebulaDatastoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	gc, err := getDatastoreController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the datastore controller",
			Detail:   err.Error(),
		})
		return diags
	}

	datastoreInfos, err := gc.Info(false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing datastore %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed retrieve datastore informations",
			Detail:   fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.Set("name", datastoreInfos.Name)
	dsTypeTpl, err := datastoreInfos.Template.Get(dsKey.Type)
	var dsType string
	for k, v := range datastoreTypes {
		if v == dsTypeTpl {
			dsType = k
		}
	}
	if err == nil {
		d.Set("type", strings.ToLower(dsType))
	}

	cfgClusterIDs := d.Get("cluster_ids").(*schema.Set).List()
	var clusterIDs []int

	if len(cfgClusterIDs) == 0 {
		// if the user hasn't configured any cluster_id
		// we ignore the the default cluster (ID: 0) at read step
		clusterIDs = make([]int, 0, len(cfgClusterIDs))

		for _, id := range datastoreInfos.Clusters.ID {
			if id == 0 {
				continue
			}
			clusterIDs = append(clusterIDs, id)
		}
	} else {
		// read all IDs
		clusterIDs = datastoreInfos.Clusters.ID
	}
	err = d.Set("cluster_ids", clusterIDs)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to set cluster_ids field",
			Detail:   fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	restrictedDirs, err := datastoreInfos.Template.Get(dsKey.RestrictedDirs)
	if err == nil {
		d.Set("restricted_directories", restrictedDirs)
	}

	safeDirs, err := datastoreInfos.Template.Get(dsKey.SafeDirs)
	if err == nil {
		d.Set("safe_directories", safeDirs)
	}

	storageUsageLimit, err := datastoreInfos.Template.Get(dsKey.LimitMB)
	if err == nil {
		d.Set("storage_usage_limit", storageUsageLimit)
	}

	transferBandwithLimit, err := datastoreInfos.Template.Get(dsKey.LimitTransferBW)
	if err == nil {
		d.Set("transfer_bandwith_limit", transferBandwithLimit)
	}

	noDecompress, err := datastoreInfos.Template.Get(dsKey.NoDecompress)
	if err == nil {
		d.Set("no_decompress", noDecompress)
	}

	checkAvailableCapacity, err := datastoreInfos.Template.Get(dsKey.DatastoreCapacityCheck)
	if err == nil {
		d.Set("check_available_capacity", checkAvailableCapacity)
	}

	bridgeList, err := datastoreInfos.Template.Get(dsKey.BridgeList)
	if err == nil {
		d.Set("bridge_list", strings.Split(bridgeList, " "))
	}

	stagingDir, err := datastoreInfos.Template.Get(dsKey.StagingDir)
	if err == nil {
		d.Set("staging_dir", stagingDir)
	}

	driver, err := datastoreInfos.Template.Get(dsKey.Driver)
	if err == nil {
		d.Set("driver", driver)
	}

	compatibleSystemDatastore, err := datastoreInfos.Template.Get(dsKey.CompatibleSysDs)
	if err == nil {
		d.Set("compatible_system_datastore", compatibleSystemDatastore)
	}

	customAttrsList := d.Get("custom").(*schema.Set).List()

	// Ceph: https://docs.opennebula.io/6.4/vmware_cluster_deployment/vmware_storage_setup/vcenter_datastores.html#vcenter-ds
	cephAttrsList := d.Get("ceph").(*schema.Set).List()

	if len(cephAttrsList) > 0 {

		cephAttrsMap := make(map[string]interface{})

		poolName, err := datastoreInfos.Template.Get("POOL_NAME")
		if err == nil {
			cephAttrsMap["pool_name"] = poolName
		}

		user, err := datastoreInfos.Template.Get("CEPH_USER")
		if err == nil {
			cephAttrsMap["user"] = user
		}

		key, err := datastoreInfos.Template.Get("CEPH_KEY")
		if err == nil {
			cephAttrsMap["key"] = key
		}

		config, err := datastoreInfos.Template.Get("CEPH_CONF")
		if err == nil {
			cephAttrsMap["config"] = config
		}

		rbdFormat, err := datastoreInfos.Template.Get("RBD_FORMAT")
		if err == nil {
			cephAttrsMap["rbd_format"] = rbdFormat
		}

		secret, err := datastoreInfos.Template.Get("CEPH_SECRET")
		if err == nil {
			cephAttrsMap["secret"] = secret
		}

		trash, err := datastoreInfos.Template.Get("CEPH_TRASH")
		if err == nil {
			cephAttrsMap["trash"] = trash
		}

		hosts, err := datastoreInfos.Template.Get("CEPH_HOST")
		if err == nil {
			cephAttrsMap["host"] = strings.Split(hosts, " ")
		}

		d.Set("ceph", []interface{}{cephAttrsMap})
	} else if len(customAttrsList) > 0 {

		customMap := map[string]interface{}{
			"transfer": datastoreInfos.TMMad,
		}

		if datastoreInfos.DSMad != "-" {
			customMap["datastore"] = datastoreInfos.DSMad
		}
		d.Set("custom", []interface{}{customMap})

	}

	flattenDiags := flattenDatastoreTemplate(d, meta, &datastoreInfos.Template)
	for _, diag := range flattenDiags {
		diags = append(diags, diag)
	}

	return nil
}

func flattenDatastoreTemplate(d *schema.ResourceData, meta interface{}, datastoreTpl *datastore.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	err := flattenTemplateSection(d, meta, &datastoreTpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to flatten template section",
			Detail:   fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err),
		})
	}

	flattenDiags := flattenTemplateTags(d, meta, &datastoreTpl.Template)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
	}

	return diags
}

func resourceOpennebulaDatastoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	dc, err := getDatastoreController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the datastore controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// template management

	datastoreInfos, err := dc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve datastores informations",
			Detail:   fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if d.HasChange("name") {
		err := dc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name for datastore %s\n", datastoreInfos.Name)
	}

	if d.HasChange("cluster_ids") {

		oldClustersIf, newClustersIf := d.GetChange("cluster_ids")

		oldClusters := schema.NewSet(schema.HashInt, oldClustersIf.(*schema.Set).List())
		newClusters := schema.NewSet(schema.HashInt, newClustersIf.(*schema.Set).List())

		// remove from clusters
		remClustersList := oldClusters.Difference(newClusters).List()

		// if the default value was set for cluster_ids (i.e. -1) at create step we remove
		// the datastore from the default cluster
		if len(oldClusters.List()) == 0 {
			for _, id := range datastoreInfos.Clusters.ID {
				if id != 0 {
					continue
				}
				remClustersList = append(remClustersList, 0)
			}
		}

		for _, id := range remClustersList {

			err = controller.Cluster(id.(int)).DelDatastore(dc.ID)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to remove from the cluster",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", dc.ID, err),
				})
				return diags
			}
		}

		// add to clusters
		addClusters := newClusters.Difference(oldClusters)

		for _, id := range addClusters.List() {
			err := controller.Cluster(id.(int)).AddDatastore(dc.ID)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add to the cluster",
					Detail:   fmt.Sprintf("cluster (ID: %d): %s", dc.ID, err),
				})
				return diags
			}
		}
	}

	update := false
	newTpl := datastoreInfos.Template

	if d.HasChange("restricted_directories") {
		restrictedDirs := d.Get("restricted_directories")
		newTpl.Del(string(dsKey.RestrictedDirs))
		newTpl.Add(dsKey.RestrictedDirs, restrictedDirs)
		update = true
	}
	if d.HasChange("safe_directories") {
		safeDirs := d.Get("safe_directories")
		newTpl.Del(string(dsKey.SafeDirs))
		newTpl.Add(dsKey.SafeDirs, safeDirs)
		update = true
	}
	if d.HasChange("no_decompress") {
		noDecompress := d.Get("no_decompress")
		newTpl.Del(string(dsKey.NoDecompress))
		newTpl.Add(dsKey.NoDecompress, noDecompress)
		update = true
	}
	if d.HasChange("storage_usage_limit") {
		storageUsageLimit := d.Get("storage_usage_limit")
		newTpl.Del(string(dsKey.LimitMB))
		newTpl.Add(dsKey.LimitMB, storageUsageLimit)
		update = true
	}
	if d.HasChange("transfer_bandwith_limit") {
		transgerBandwithLimit := d.Get("transfer_bandwith_limit")
		newTpl.Del(string(dsKey.LimitTransferBW))
		newTpl.Add(dsKey.LimitTransferBW, transgerBandwithLimit)
		update = true
	}
	if d.HasChange("check_available_capacity") {
		checkAvailableCapacity := d.Get("check_available_capacity")
		newTpl.Del(string(dsKey.DatastoreCapacityCheck))
		newTpl.Add(dsKey.DatastoreCapacityCheck, checkAvailableCapacity)
		update = true
	}
	if d.HasChange("bridge_list") {
		brigeList := d.Get("bridge_list")
		newTpl.Del(string(dsKey.BridgeList))
		newTpl.Add(dsKey.BridgeList, brigeList)
		update = true
	}
	if d.HasChange("staging_dir") {
		stagingDir := d.Get("staging_dir")
		newTpl.Del(string(dsKey.StagingDir))
		newTpl.Add(dsKey.StagingDir, stagingDir)
		update = true
	}
	if d.HasChange("driver") {
		driver := d.Get("driver")
		newTpl.Del(string(dsKey.Driver))
		newTpl.Add(dsKey.Driver, driver)
		update = true
	}
	if d.HasChange("compatible_system_datastore") {
		compatibleSystemDS := d.Get("compatible_system_datastore")
		newTpl.Del(string(dsKey.CompatibleSysDs))
		newTpl.Add(dsKey.CompatibleSysDs, compatibleSystemDS)
		update = true
	}

	if d.HasChange("ceph") {
		cephAttrsList := d.Get("ceph").(*schema.Set).List()
		addCephAttributes(cephAttrsList[0].(map[string]interface{}), &newTpl)
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
		err = dc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update datastore content",
				Detail:   fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

	}

	return resourceOpennebulaDatastoreRead(ctx, d, meta)
}

func resourceOpennebulaDatastoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	dc, err := getDatastoreController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the datastore controller",
			Detail:   err.Error(),
		})
		return diags
	}
	err = dc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("datastore (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return nil
}
