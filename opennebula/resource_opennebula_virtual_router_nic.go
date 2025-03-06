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

var defaultVRNICTimeout = time.Duration(10) * time.Minute

func resourceOpennebulaVirtualRouterNIC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualRouterNICCreate,
		ReadContext:   resourceOpennebulaVirtualRouterNICRead,
		Exists:        resourceOpennebulaVirtualRouterNICExists,
		DeleteContext: resourceOpennebulaVirtualRouterNICDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultVRNICTimeout),
			Delete: schema.DefaultTimeout(defaultVRNICTimeout),
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
			"floating_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"floating_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"ip6": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
	if v, ok := d.GetOk("floating_ip"); ok {
		if v.(bool) {
			nicTpl.Add("FLOATING_IP", "YES")
		}
	}
	if v, ok := d.GetOk("floating_only"); ok {
		if v.(bool) {
			nicTpl.Add("FLOATING_ONLY", "YES")
		}
	}

	if v, ok := d.GetOk("ip"); ok {
		nicTpl.Add("IP", v.(string))
	}

	if v, ok := d.GetOk("ip6"); ok {
		nicTpl.Add("IP6", v.(string))
	}

	// wait before checking NIC
	nicID, err := vrNICAttach(ctx, d.Timeout(schema.TimeoutCreate), controller, vRouterID, nicTpl)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to attach NIC",
			Detail:   fmt.Sprintf("virtual router NIC (ID: %s): %s", d.Id(), err),
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
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual router NIC (ID: %s): %s", d.Id(), err),
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
			Summary:  "Failed to find the NIC in the virtual router NIC list",
			Detail:   fmt.Sprintf("virtual router NIC (ID: %s): %s", d.Id(), err),
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

	floatingIP, _ := nic.GetStr("FLOATING_IP")
	floatingOnly, _ := nic.GetStr("FLOATING_ONLY")
	ip, _ := nic.Get("IP")
	ip6, _ := nic.Get("IP6")

	// For VRouter NICs, floating IPs are set using the "IP" field, but it is represented in the ON API
	// as the "VROUTER_IP" field.
	if strings.ToUpper(floatingIP) == "YES" {
		if vrouterIp, _ := nic.Get("VROUTER_IP"); len(vrouterIp) > 0 {
			ip = vrouterIp
		}
		if vrouterIp6, _ := nic.Get("VROUTER_IP6"); len(vrouterIp6) > 0 {
			ip6 = vrouterIp6
		}
	}

	d.Set("network_id", networkID)
	d.Set("virtual_router_id", vr.ID)
	d.Set("physical_device", phyDev)
	d.Set("network", network)
	d.Set("model", model)
	d.Set("virtio_queues", virtioQueues)
	d.Set("security_groups", sg)
	d.Set("floating_ip", strings.ToUpper(floatingIP) == "YES")
	d.Set("floating_only", strings.ToUpper(floatingOnly) == "YES")
	d.Set("ip", ip)
	d.Set("ip6", ip6)

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
			Summary:  "Failed to parse virtual router ID",
			Detail:   fmt.Sprintf("virtual router NIC (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// wait before checking NIC
	err = vrNICDetach(ctx, d.Timeout(schema.TimeoutCreate), controller, vRouterID, int(nicID))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to detach NIC",
			Detail:   fmt.Sprintf("virtual router NIC (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully detached virtual router NIC\n")
	return nil
}
