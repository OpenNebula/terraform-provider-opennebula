package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/host"
	hostk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/host/keys"
)

var hostTypes = []string{"kvm", "qemu", "lxd", "lxc", "firecracker", "vcenter", "custom"}
var defaultHostMinTimeout = 20
var defaultHostTimeout = time.Duration(defaultHostMinTimeout) * time.Minute

func resourceOpennebulaHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaHostCreate,
		ReadContext:   resourceOpennebulaHostRead,
		UpdateContext: resourceOpennebulaHostUpdate,
		DeleteContext: resourceOpennebulaHostDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultHostTimeout),
			Delete: schema.DefaultTimeout(defaultHostTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Host",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Type of the new host: kvm, qemu, lxd, lxc, firecracker, custom",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {

					if inArray(v.(string), hostTypes) < 0 {
						errors = append(errors, fmt.Errorf("host \"type\" must be one of: %s", strings.Join(hostTypes, ",")))
					}

					return
				},
			},
			"custom": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"virtualization": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Virtualization driver",
						},
						"information": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Information driver",
						},
					},
				},
			},
			"overcommit": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum allocatable CPU capacity in number of cores multiplied by 100",
						},
						"memory": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Maximum allocatable memory capacity in KB",
						},
					},
				},
			},
			"cluster_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     -1,
				Description: "ID of the cluster. Affected to the default cluster if not set",
			},
			"tags":         tagsSchema(),
			"default_tags": defaultTagsSchemaComputed(),
			"tags_all":     tagsSchemaComputed(),
		},
	}
}

func getHostController(d *schema.ResourceData, meta interface{}) (*goca.HostController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	uid, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.Host(int(uid)), nil
}

func resourceOpennebulaHostCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	hostType := d.Get("type").(string)

	var vmMad, imMad string

	switch hostType {
	case "kvm", "qemu", "lxd", "lxc", "firecracker":
		imMad = hostType
		vmMad = hostType

	case "vcenter":
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "The vcenter type is not managed.",
		})
		return diags
	case "custom":

		madsList := d.Get("custom").(*schema.Set).List()

		for _, madsIf := range madsList {
			madsMap := madsIf.(map[string]interface{})
			imMadIf, ok := madsMap["information"]
			if !ok {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "No information field found in the custom section",
				})
				return diags
			}
			imMad = imMadIf.(string)
			vmMadIf, ok := madsMap["virtualization"]
			if !ok {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "No virtualization field found in the custom section",
				})
				return diags
			}
			vmMad = vmMadIf.(string)
		}

	}

	clusterID := d.Get("cluster_id").(int)

	log.Printf("[INFO] Host %s %s %s", name, imMad, vmMad)

	hostID, err := controller.Hosts().Create(name, imMad, vmMad, clusterID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the host",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%v", hostID))

	log.Printf("[INFO] Host created")

	hc := controller.Host(hostID)

	timeout := d.Timeout(schema.TimeoutCreate)
	_, err = waitForHostStates(ctx, hc, timeout, []string{"INIT", "MONITORING_INIT", "MONITORING_MONITORED"}, []string{"MONITORED", "INIT"})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to wait host to be in MONITORED state",
			Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	hostTpl, err := generateHostTemplate(d, meta, hc)

	hostTplStr := hostTpl.String()
	log.Printf("[INFO] Host template: %s", hostTplStr)

	err = hc.Update(hostTplStr, parameters.Replace)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve information",
			Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	return resourceOpennebulaHostRead(ctx, d, meta)
}

func generateHostTemplate(d *schema.ResourceData, meta interface{}, hc *goca.HostController) (*dyn.Template, error) {

	config := meta.(*Configuration)
	tpl := dyn.NewTemplate()

	err := generateHostOvercommit(d, meta, hc, tpl)
	if err != nil {
		return nil, err
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

	return tpl, nil
}

func generateHostOvercommit(d *schema.ResourceData, meta interface{}, hc *goca.HostController, tpl *dyn.Template) error {

	overcommit, ok := d.GetOk("overcommit")
	if ok {
		overcommitList := overcommit.(*schema.Set).List()

		hostInfos, err := hc.Info(false)
		if err != nil {
			return fmt.Errorf("Failed to retrieve informations")
		}

		for _, overcommitIf := range overcommitList {
			overcommitMap := overcommitIf.(map[string]interface{})
			cpuIf, ok := overcommitMap["cpu"]
			if !ok {
				return fmt.Errorf("No cpu field found in the overcommit section")
			}
			memoryIf, ok := overcommitMap["memory"]
			if !ok {
				return fmt.Errorf("No memory field found in the overcommit section")
			}

			reservedCPU := hostInfos.Share.TotalCPU - cpuIf.(int)
			reservedMem := hostInfos.Share.TotalMem - memoryIf.(int)

			tpl.AddPair("RESERVED_CPU", reservedCPU)
			tpl.AddPair("RESERVED_MEM", reservedMem)
		}
	}

	return nil
}

func resourceOpennebulaHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	hc, err := getHostController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the host controller",
			Detail:   err.Error(),
		})
		return diags

	}

	hostInfos, err := hc.Info(false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing host %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", hostInfos.ID))
	d.Set("name", hostInfos.Name)

	custom := d.Get("custom").(*schema.Set).List()
	if len(custom) > 0 {

		d.Set("custom", []map[string]interface{}{
			{
				"information":    hostInfos.IMMAD,
				"virtualization": hostInfos.VMMAD,
			},
		})
	}

	overcommit := d.Get("overcommit").(*schema.Set).List()
	if len(overcommit) > 0 {
		overcommitMap := make(map[string]interface{})

		reservedCPU, err := hostInfos.Template.GetI(hostk.ReservedCPU)
		if err == nil {
			overcommitMap["cpu"] = hostInfos.Share.TotalCPU - reservedCPU
		}

		reservedMem, err := hostInfos.Template.GetInt("RESERVED_MEM")
		if err == nil {
			overcommitMap["memory"] = hostInfos.Share.TotalMem - reservedMem
		}

		if len(overcommitMap) > 0 {
			d.Set("overcommit", []map[string]interface{}{overcommitMap})
		}
	}

	clusterID := d.Get("cluster_id").(int)
	if clusterID != -1 {

		d.Set("cluster_id", hostInfos.ClusterID)
	}

	flattenDiags := flattenHostTemplate(d, meta, &hostInfos.Template)
	for _, diag := range flattenDiags {
		diags = append(diags, diag)
	}

	return diags
}

func flattenHostTemplate(d *schema.ResourceData, meta interface{}, hostTpl *host.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	err := flattenTemplateSection(d, meta, &hostTpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Failed to read template section",
			Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
		})
	}

	flattenDiags := flattenTemplateTags(d, meta, &hostTpl.Template)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("host (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
	}

	return diags
}

func resourceOpennebulaHostUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	config := meta.(*Configuration)
	controller := config.Controller

	hc, err := getHostController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the host controller",
			Detail:   err.Error(),
		})
		return diags
	}

	hostInfos, err := hc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if d.HasChange("name") {
		err := hc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name for host %s\n", hostInfos.Name)
	}

	if d.HasChange("cluster_id") {
		clusterID := d.Get("cluster_id").(int)

		err := controller.Cluster(clusterID).AddHost(hc.ID)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to add host to it's new cluster",
				Detail:   fmt.Sprintf("host (ID: %s) cluster (ID: %d): %s", d.Id(), clusterID, err),
			})
			return diags
		}
	}

	update := false
	newTpl := hostInfos.Template

	if d.HasChange("overcommit") {
		newTpl.Del("RESERVED_CPU")
		newTpl.Del("RESERVED_MEM")
		err := generateHostOvercommit(d, meta, hc, &newTpl.Template)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to compute host overcommit",
				Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
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
			newTpl.Del(strings.ToUpper(k))
			newTpl.AddPair(strings.ToUpper(k), v)
		}

		update = true
	}

	if update {
		err = hc.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaHostRead(ctx, d, meta)
}

func resourceOpennebulaHostDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	hc, err := getHostController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the host controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = hc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete the host",
			Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	_, err = waitForHostStates(ctx, hc, timeout, []string{"INIT", "MONITORING_INIT", "MONITORING_MONITORED", "MONITORED"}, []string{"notfound"})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to wait host to be in notfound state",
			Detail:   fmt.Sprintf("host (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	return nil
}

// waitForHostStates wait for a an host to reach some expected states
func waitForHostStates(ctx context.Context, hc *goca.HostController, timeout time.Duration, pending, target []string) (interface{}, error) {

	stateChangeConf := resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
		Refresh: func() (interface{}, string, error) {

			log.Println("Refreshing host state...")

			hostInfos, err := hc.Info(false)
			if err != nil {
				if NoExists(err) {
					return hostInfos, "notfound", nil
				}
				return hostInfos, "", err
			}
			state, err := hostInfos.State()
			if err != nil {
				return hostInfos, "", err
			}

			log.Printf("Host (ID:%d, name:%s) is currently in state %s", hostInfos.ID, hostInfos.Name, state.String())

			// In case we are in some failure state, we try to retrieve more error informations from the host template
			if state == host.Error {
				hostErr, _ := hostInfos.Template.Get("ERROR")
				return hostInfos, state.String(), fmt.Errorf("Host (ID:%d) entered fail state, error: %s", hostInfos.ID, hostErr)
			}

			return hostInfos, state.String(), nil
		},
	}

	return stateChangeConf.WaitForStateContext(ctx)

}
