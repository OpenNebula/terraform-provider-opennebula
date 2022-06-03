package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func resourceOpennebulaVirtualRouterNIC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualRouterNICCreate,
		ReadContext:   resourceOpennebulaVirtualRouterNICRead,
		Exists:        resourceOpennebulaVirtualRouterNICExists,
		DeleteContext: resourceOpennebulaVirtualRouterNICDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"virtual_router_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			// Following fields are similar to those from nicFields method
			// except some additional behavior Computer and ForceNew
			"model": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"virtio_queues": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Only if model is virtio",
			},
			"network_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"physical_device": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func resourceOpennebulaVirtualRouterNICCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller

	vRouterID := d.Get("virtual_router_id").(int)

	// avoid creation of multiple NICs and instances at the same time
	nicKey := &SubResourceKey{
		Type:    "virtual_router",
		ID:      vRouterID,
		SubType: "nic",
	}
	config.mutex.Lock(nicKey)
	defer config.mutex.Unlock(nicKey)

	nicTpl := shared.NewNIC()
	vnetID := d.Get("network_id").(int)

	nicTpl.Add(shared.NetworkID, vnetID)

	if v, ok := d.GetOk("model"); ok {
		nicTpl.Add(shared.Model, v.(string))
	}
	if v, ok := d.GetOk("virtio_queues"); ok {
		nicTpl.Add("VIRTIO_QUEUES", v.(string))
	}
	if v, ok := d.GetOk("physical_device"); ok {
		nicTpl.Add("PHYDEV", v.(string))
	}
	if v, ok := d.GetOk("security_groups"); ok {
		secGroups := ArrayToString(v.([]interface{}), ",")
		nicTpl.Add(shared.SecurityGroups, secGroups)
	}

	// wait before checking NIC
	nicID, err := vrNICAttach(ctx, d.Timeout(schema.TimeoutCreate), controller, vRouterID, nicTpl)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router nic attach failed",
			Detail:   fmt.Sprintf("Virtual router (ID:%d): %s", vRouterID, err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%d", nicID))

	log.Printf("[INFO] Successfully attached virtual router NIC\n")

	return resourceOpennebulaVirtualRouterNICRead(ctx, d, meta)
}

func resourceOpennebulaVirtualRouterNICRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller
	vrouterID := d.Get("virtual_router_id").(int)

	vr, err := controller.VirtualRouter(vrouterID).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "get virtual router informations",
			Detail:   fmt.Sprintf("Virtual router (ID:%d): %s", vrouterID, err),
		})
		return diags
	}

	// get the nic ID from the nic list
	var nic *shared.NIC

	nics := vr.Template.GetNICs()
	for _, n := range nics {
		nicID, _ := n.Get(shared.NICID)
		if nicID == d.Id() {
			nic = &n
			break
		}
	}

	if nic == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router nic not found",
			Detail:   fmt.Sprintf("Virtual router (ID:%d): NIC (ID:%s) not found", vrouterID, d.Id()),
		})
		return diags
	}

	networkID, _ := nic.GetI(shared.NetworkID)
	phyDev, _ := nic.GetStr("PHYDEV")
	network, _ := nic.Get(shared.Network)
	model, _ := nic.Get(shared.Model)
	virtioQueues, _ := nic.GetStr("VIRTIO_QUEUES")

	sg := make([]int, 0)
	securityGroupsArray, _ := nic.Get(shared.SecurityGroups)
	sgString := strings.Split(securityGroupsArray, ",")
	for _, s := range sgString {
		sgInt, _ := strconv.ParseInt(s, 10, 32)
		sg = append(sg, int(sgInt))
	}

	d.Set("network_id", networkID)
	d.Set("virtual_router_id", vr.ID)
	d.Set("physical_device", phyDev)
	d.Set("network", network)
	d.Set("model", model)
	d.Set("virtio_queues", virtioQueues)
	d.Set("security_groups", sg)

	return nil
}

func resourceOpennebulaVirtualRouterNICExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	vrouterID := d.Get("virtual_router_id").(int)

	_, err := controller.VirtualRouter(vrouterID).Info(false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaVirtualRouterNICDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller
	vRouterID := d.Get("virtual_router_id").(int)

	// avoid creation of multiple NICs and instances at the same time
	nicKey := &SubResourceKey{
		Type:    "virtual_router",
		ID:      vRouterID,
		SubType: "nic",
	}
	config.mutex.Lock(nicKey)
	defer config.mutex.Unlock(nicKey)

	nicID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "can't parse virtual router ID",
			Detail:   fmt.Sprintf("%s is not an ID: %s", d.Id(), err),
		})
		return diags
	}

	// wait before checking NIC
	err = vrNICDetach(ctx, d.Timeout(schema.TimeoutCreate), controller, vRouterID, int(nicID))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "can't detach virtual router NIC",
			Detail:   fmt.Sprintf("Virtual router (ID:%d): NIC (ID:%d): %s", vRouterID, int(nicID), err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully detached virtual router NIC\n")
	return nil
}
