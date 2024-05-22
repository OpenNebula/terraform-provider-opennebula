package opennebula

import (
	"context"
	"fmt"
	"log"
	// "strings"

	// vn "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Define the schema for the data source.
func dataSourceOpennebulaVirtualNetworkAddressRanges() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOpennebulaVirtualNetworkAddressRangesRead,
		Schema: map[string]*schema.Schema{
			"virtual_network_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"address_ranges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the address range.",
						},
						"ar_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the Address Range: IP4, IP6.",
						},
						"ip4": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Start IPv4 of the range to be allocated.",
						},
						"size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Count of addresses in the IP range.",
						},
					},
				},
			},
		},
	}
}

// Read the address ranges of a virtual network.
func dataSourceOpennebulaVirtualNetworkAddressRangesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Iterate over each address range and extract minimal information.
	for _, addressRange := range virtualNetworkInfo.ARs {
		minimalInfo := map[string]interface{}{
			"id":      addressRange.ID,
			"ar_type": addressRange.Type,
			"ip4":     addressRange.IP,
			"size":    addressRange.Size,
		}
		addressRanges = append(addressRanges, minimalInfo)
	}

	// Set the result and log success.
	d.Set("address_ranges", addressRanges)
	d.SetId(fmt.Sprintf("%d", virtualNetworkID))
	log.Printf("[INFO] Successfully retrieved address ranges of the virtual network\n")

	return diags
}
