package opennebula

import (
	"context"
	"fmt"
	"log"
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
			},
			"address_range_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"address_ranges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ar_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the Address Range: IP4, IP6. Default is 'IP4'",
						},
						"ip4": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Start IPv4 of the range to be allocated (Required if IP4 or IP4_6).",
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
						"global_prefix": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Global prefix for IP6 or IP4_6",
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

	// Retrieve information about address ranges for the specified virtual network.
	virtualNetworkInfo, err := controller.VirtualNetwork(virtualNetworkID).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve information about the virtual network",
			Detail:   fmt.Sprintf("Virtual Network (ID:%d): %s", virtualNetworkID, err),
		})
		return diags
	}

	// Prepare the results array.
	var addressRanges []interface{}

	// If address_range_id is provided, retrieve only that specific address range.
	if addressRangeID, ok := d.Get("address_range_id").(string); ok && addressRangeID != "" {
		for _, addressRange := range virtualNetworkInfo.ARs {
			if addressRange.ID == addressRangeID {
				addressRanges = append(addressRanges, flattenAddressRange(addressRange))
				break
			}
		}
	} else {
		// If address_range_id is not provided, retrieve all address ranges.
		for _, addressRange := range virtualNetworkInfo.ARs {
			addressRanges = append(addressRanges, flattenAddressRange(addressRange))
		}
	}

	// Set the result and log success.
	d.Set("address_ranges", addressRanges)
	d.SetId(fmt.Sprintf("%d", virtualNetworkID))
	log.Printf("[INFO] Successfully retrieved address ranges of the virtual network\n")

	return diags
}

// Flatten the structure of a virtual network address range into a usable form.
func flattenAddressRange(ar vn.AR) map[string]interface{} {
	addressRange := make(map[string]interface{})

	// Populate fields.
	addressRange["ar_type"] = ar.Type
	addressRange["ip4"] = ar.IP
	addressRange["size"] = ar.Size
	addressRange["mac"] = ar.MAC
	addressRange["global_prefix"] = ar.GlobalPrefix

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
