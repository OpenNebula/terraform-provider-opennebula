package opennebula

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	vn "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"
	vnk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork/keys"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpennebulaVirtualNetworkAddressRange() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualNetworkAddressRangeCreate,
		ReadContext:   resourceOpennebulaVirtualNetworkAddressRangeRead,
		UpdateContext: resourceOpennebulaVirtualNetworkAddressRangeUpdate,
		Exists:        resourceOpennebulaVirtualNetworkAddressRangeExists,
		DeleteContext: resourceOpennebulaVirtualNetworkAddressRangeDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"virtual_network_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"ar_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "IP4",
				Description: "Type of the Address Range: IP4, IP6. Default is 'IP4'",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					validtypes := []string{"IP4", "IP6", "IP6_STATIC", "IP4_6", "IP4_6_STATIC", "ETHER"}
					value := v.(string)

					if inArray(value, validtypes) < 0 {
						errors = append(errors, fmt.Errorf("Address Range type %q must be one of: %s", k, strings.Join(validtypes, ",")))
					}

					return
				},
			},
			"ip4": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Start IPv4 of the range to be allocated (Required if IP4 or IP4_6).",
			},
			"size": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Size (in number) of the ip range",
			},
			"ip6": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Start IPv6 of the range to be allocated (Required if IP6_STATIC or IP4_6_STATIC)",
			},
			"mac": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Start MAC of the range to be allocated",
			},
			"global_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Global prefix for IP6 or IP4_6",
			},
			"ula_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ULA prefix for IP6 or IP4_6",
			},
			"prefix_length": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Prefix lenght Only needed for IP6_STATIC or IP4_6_STATIC",
			},
			"hold_ips": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of IPs to be held the VNET",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceOpennebulaVirtualNetworkAddressRangeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller

	vNetworkID := d.Get("virtual_network_id").(int)

	vnc := controller.VirtualNetwork(vNetworkID)

	arTpl := generateAR(d)
	arID, err := vNetARAdd(ctx, d.Timeout(schema.TimeoutCreate), vnc, vNetworkID, arTpl)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual network address range adding failed",
			Detail:   fmt.Sprintf("Virtual network (ID:%d): %s", vNetworkID, err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%d", arID))

	if holdIPs, ok := d.GetOk("hold_ips"); ok {
		for _, ip := range holdIPs.([]interface{}) {
			err = ipHold(vnc, ip.(string))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to hold a lease",
					Detail:   fmt.Sprintf("Virtual network (ID: %d) address range (ID: %s): %s", vNetworkID, d.Id(), err),
				})
				return diags
			}
		}
	}

	log.Printf("[INFO] Successfully added virtual network AR\n")

	return resourceOpennebulaVirtualNetworkAddressRangeRead(ctx, d, meta)
}

func generateAR(d *schema.ResourceData) *vn.AddressRange {

	ar := vn.NewAddressRange()

	// Generate AR depending on the AR Type
	artype := d.Get("ar_type").(string)
	arip4 := d.Get("ip4").(string)
	arip6 := d.Get("ip6").(string)
	armac := d.Get("mac").(string)
	arsize := d.Get("size").(int)
	argprefix := d.Get("global_prefix").(string)
	arulaprefix := d.Get("ula_prefix").(string)
	arprefixlength := d.Get("prefix_length").(string)

	ar.Add(vnk.Size, fmt.Sprint(arsize))
	ar.Add(vnk.Type, artype)

	if armac != "" {
		ar.Add(vnk.Mac, armac)
	}

	switch artype {
	case "IP4":
		ar.Add(vnk.IP, arip4)

	case "IP6":

		if argprefix != "" {
			ar.Add(vnk.GlobalPrefix, argprefix)
		}

		if arulaprefix != "" {
			ar.Add(vnk.UlaPrefix, arulaprefix)
		}

	case "IP6_STATIC":

		ar.Add("IP6", arip6)
		ar.Add(vnk.PrefixLength, arprefixlength)

	case "IP4_6":

		if argprefix != "" {
			ar.Add(vnk.GlobalPrefix, argprefix)
		}

		if arulaprefix != "" {
			ar.Add(vnk.UlaPrefix, arulaprefix)
		}

		ar.Add(vnk.IP, arip4)

	case "IP4_6_STATIC":

		ar.Add(vnk.IP, arip4)
		ar.Add("IP6", arip6)
		ar.Add(vnk.PrefixLength, arprefixlength)
	}

	return ar
}

func resourceOpennebulaVirtualNetworkAddressRangeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller
	vNetworkID := d.Get("virtual_network_id").(int)

	vnInfos, err := controller.VirtualNetwork(vNetworkID).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "get virtual network informations",
			Detail:   fmt.Sprintf("Virtual network (ID:%d): %s", vNetworkID, err),
		})
		return diags
	}

	// match the AR in the list
	var ar *vn.AR

	ars := vnInfos.ARs
	for i, a := range ars {
		if a.ID == d.Id() {
			ar = &ars[i]
			break
		}
	}

	if ar == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual network address range not found",
			Detail:   fmt.Sprintf("Virtual network (ID:%d): AR (ID:%s) not found", vNetworkID, d.Id()),
		})
		return diags
	}

	d.Set("ar_type", ar.Type)
	d.Set("ip4", ar.IP)
	d.Set("ip6", ar.IP6)
	d.Set("size", ar.Size)
	d.Set("mac", ar.MAC)
	d.Set("ula_prefix", ar.GlobalPrefix)

	leases := make([]string, 0, len(ar.Leases))
	for _, lease := range ar.Leases {
		leases = append(leases, lease.IP)
	}
	d.Set("hold_ips", leases)

	return nil
}

func resourceOpennebulaVirtualNetworkAddressRangeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller

	vNetworkID := d.Get("virtual_network_id").(int)
	vnc := controller.VirtualNetwork(vNetworkID)

	if d.HasChange("hold_ips") {
		oldIPs, newIPs := d.GetChange("hold_ips")

		oldIPsSet := schema.NewSet(schema.HashString, oldIPs.([]interface{}))
		newIPsSet := schema.NewSet(schema.HashString, newIPs.([]interface{}))

		// release some old IPs
		remIPs := oldIPsSet.Difference(newIPsSet)
		for _, ip := range remIPs.List() {

			// release the IP
			err := ipRelease(vnc, ip.(string))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to release a lease on hold",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		// hold some new IPs
		addIPs := newIPsSet.Difference(oldIPsSet)
		for _, ip := range addIPs.List() {

			// release the IP
			err := ipHold(vnc, ip.(string))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to release a lease on hold",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

	} else {
		arTpl := generateAR(d)
		err := vnc.UpdateAR(arTpl.String())
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "virtual network address range adding failed",
				Detail:   fmt.Sprintf("Virtual network (ID:%d): %s", vNetworkID, err),
			})
			return diags
		}
	}

	log.Printf("[INFO] Successfully updated virtual network AR\n")

	return resourceOpennebulaVirtualNetworkAddressRangeRead(ctx, d, meta)
}

func resourceOpennebulaVirtualNetworkAddressRangeExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	vrouterID := d.Get("virtual_network_id").(int)

	_, err := controller.VirtualNetwork(vrouterID).Info(false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaVirtualNetworkAddressRangeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller
	vNetworkID := d.Get("virtual_network_id").(int)
	vnc := controller.VirtualNetwork(vNetworkID)

	if holdIPs, ok := d.GetOk("hold_ips"); ok {

		for _, ip := range holdIPs.([]interface{}) {

			err := ipRelease(vnc, ip.(string))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to release a lease on hold",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

		}
	}
	log.Printf("[INFO] Successfully released reservered IP addresses.")

	arID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "can't parse virtual network ID",
			Detail:   fmt.Sprintf("%s is not an ID: %s", d.Id(), err),
		})
		return diags
	}

	err = vNetARDetach(ctx, d.Timeout(schema.TimeoutCreate), controller, vNetworkID, int(arID))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "can't detach virtual network AR",
			Detail:   fmt.Sprintf("Virtual network (ID:%d): AR (ID:%d): %s", vNetworkID, int(arID), err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully detached virtual network AR\n")
	return nil
}

func ipHold(vnc *goca.VirtualNetworkController, ip string) error {

	addressReservation := dynamic.Vector{
		XMLName: xml.Name{Local: "LEASES"},
	}
	addressReservation.AddPair("IP", ip)

	err := vnc.Hold(addressReservation.String())
	if err != nil {
		return err
	}

	return nil
}

func ipRelease(vnc *goca.VirtualNetworkController, ip string) error {
	addressReservation := dynamic.Vector{
		XMLName: xml.Name{Local: "LEASES"},
	}
	addressReservation.AddPair("IP", ip)

	err := vnc.Release(addressReservation.String())
	if err != nil {
		return err
	}

	return nil
}
