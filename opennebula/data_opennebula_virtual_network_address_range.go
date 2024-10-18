package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	vn "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Define the schema for the data source.
func dataSourceOpennebulaVirtualNetworkAddressRange() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpennebulaVirtualNetworkAddressRangeRead,
		Schema: map[string]*schema.Schema{
			"virtual_network_id": {
				Type:     schema.TypeInt,
				Required: true,
				Description: "Id of the virtual network",
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Id of the virtual network range",
			},
			"ar_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the Address Range: IP4, IP6, IP4_6",
			},
			"ip4": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Start IPv4 of the range to be allocated",
			},
			"ip4_end": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "End IPv4 of the range to be allocated",
			},
			"ip6": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Start IPv6 of the range to be allocated",
			},
			"ip6_end": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "End IPv6 of the range to be allocated",
			},
			"ip6_global": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Global IPv6 of the range to be allocated",
			},
			"ip6_global_end": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "End Global IPv6 of the range to be allocated",
			},
			"ip6_ula": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ULA IPv6 of the range to be allocated",
			},
			"ip6_ula_end": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "End ULA IPv6 of the range to be allocated",
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Count of addresses in the IP range",
			},
			"mac": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Start MAC of the range to be allocated",
			},
			"mac_end": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "End MAC of the range to be allocated",
			},
			"global_prefix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Global prefix for IP6 or IP4_6",
			},
			"ula_prefix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ULA prefix for IP6 or IP4_6",
			},
			"held_ips": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of IPs held in this address range",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"custom": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// Read the address range of a virtual network.
func dataSourceOpennebulaVirtualNetworkAddressRangeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get configuration from meta.
	config := meta.(*Configuration)
	controller := config.Controller
	virtualNetworkID := d.Get("virtual_network_id").(int)
	addressRangeIDStr := d.Get("id").(string)

	// Convert addressRangeIDStr to an integer.
	addressRangeID, err := strconv.Atoi(addressRangeIDStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid address range ID",
			Detail:   fmt.Sprintf("Address Range ID should be an integer, got: %s", addressRangeIDStr),
		})
		return diags
	}

	// Retrieve information about the virtual network.
	virtualNetworkInfo, err := controller.VirtualNetwork(virtualNetworkID).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve information about the virtual network",
			Detail:   fmt.Sprintf("Virtual Network (ID:%d): %s", virtualNetworkID, err),
		})
		return diags
	}

	// Validate the address range ID.
	if addressRangeID < 0 || addressRangeID >= len(virtualNetworkInfo.ARs) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Address Range ID out of bounds",
			Detail:   fmt.Sprintf("Address Range ID (ID:%d) is out of bounds for Virtual Network (ID:%d)", addressRangeID, virtualNetworkID),
		})
		return diags
	}

	// Get the specified address range.
	addressRange := virtualNetworkInfo.ARs[addressRangeID]

	// Flatten the address range and set the fields.
	flattenedAddressRange := flattenAddressRange(addressRange)
	for key, value := range flattenedAddressRange {
		d.Set(key, value)
	}

	// Set the ID of the resource.
	d.SetId(fmt.Sprintf("%d", addressRangeID))
	log.Printf("[INFO] Successfully retrieved address range of the virtual network\n")

	return diags
}

// Flatten the structure of a virtual network address range into a usable form.
func flattenAddressRange(ar vn.AR) map[string]interface{} {
	addressRange := make(map[string]interface{})

	// Populate fields.
	addressRange["ar_type"] = ar.Type
	addressRange["ip4"] = ar.IP
	addressRange["ip4_end"] = ar.IPEnd
	addressRange["ip6"] = ar.IP6
	addressRange["ip6_end"] = ar.IP6End
	addressRange["ip6_global"] = ar.IP6Global
	addressRange["ip6_global_end"] = ar.IP6GlobalEnd
	addressRange["ip6_ula"] = ar.IP6ULA
	addressRange["ip6_ula_end"] = ar.IP6ULAEnd
	addressRange["size"] = ar.Size
	addressRange["mac"] = ar.MAC
	addressRange["mac_end"] = ar.MACEnd
	addressRange["global_prefix"] = ar.GlobalPrefix
	addressRange["ula_prefix"] = ar.ULAPrefix

	// Flatten held IPs.
	heldIPs := make([]interface{}, len(ar.Leases))
	for i, lease := range ar.Leases {
		heldIPs[i] = lease.IP
	}
	addressRange["held_ips"] = heldIPs

	// Flatten custom attributes.
	custom := make(map[string]interface{})
	for _, pair := range ar.Custom {
		custom[strings.ToLower(pair.Key())] = pair.Value
	}
	addressRange["custom"] = custom

	return addressRange
}
