package opennebula

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceOpennebulaVirtualNetworkAddressRanges(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualNetworkAddressRangesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVirtualNetworkAddressRangesConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("opennebula_virtual_network_address_range.test_ar_1", "id"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network_address_range.test_ar_2", "id"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network_address_range.test_ar_3", "id"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network_address_range.test_ar_4", "id"),
				),
			},
			{
				Config: testAccDataSourceVirtualNetworkAddressRangesConfig + testAccDataSourceVirtualNetworkAddressRanges,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.opennebula_virtual_network_address_ranges.test_ar", "virtual_network_id",
						"opennebula_virtual_network.test_vnet_address_ranges", "id",
					),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_ranges.test_ar", "address_ranges.#", "4"),

					resource.TestCheckTypeSetElemNestedAttrs(
						"data.opennebula_virtual_network_address_ranges.test_ar", "address_ranges.*", map[string]string{
							"ar_type": "IP4",
							"ip4":     "172.16.100.110",
							"size":    "16",
							"mac":     "02:00:00:00:00:01",
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(
						"data.opennebula_virtual_network_address_ranges.test_ar", "address_ranges.*", map[string]string{
							"ar_type": "IP6_STATIC",
							"ip6":     "2001:db8::1",
							"size":    "15",
							"mac":     "02:00:00:00:01:01",
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(
						"data.opennebula_virtual_network_address_ranges.test_ar", "address_ranges.*", map[string]string{
							"ar_type": "IP4",
							"ip4":     "172.16.150.100",
							"size":    "32",
						},
					),

					resource.TestCheckTypeSetElemNestedAttrs(
						"data.opennebula_virtual_network_address_ranges.test_ar", "address_ranges.*", map[string]string{
							"ar_type": "IP4",
							"ip4":     "192.168.0.200",
							"size":    "3",
						},
					),
				),
			},
		},
	})
}

func testAccCheckVirtualNetworkAddressRangesDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "opennebula_virtual_network" && rs.Primary.Attributes["name"] == "test_vnet_address_ranges" {
			vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			vnc := controller.VirtualNetwork(int(vnID))

			// Wait for Virtual Network deleted
			stateConf := &resource.StateChangeConf{
				Pending: []string{"exists"},
				Target:  []string{"deleted"},
				Refresh: func() (any, string, error) {

					vn, _ := vnc.Info(false)
					if vn == nil {
						return vn, "deleted", nil
					}

					return vn, "exists", nil
				},
				Timeout:    1 * time.Minute,
				Delay:      10 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			_, err := stateConf.WaitForStateContext(context.Background())

			return err
		}
	}

	return nil
}

var testAccDataSourceVirtualNetworkAddressRangesConfig = `
resource "opennebula_virtual_network" "test_vnet_address_ranges" {
    name            = "test-vnet-addresses"
    type            = "dummy"
    bridge          = "onebr"
    mtu             = 1500
    gateway         = "172.16.100.1"
    dns             = "172.16.100.1"
    network_mask    = "255.255.255.0"
    network_address = "172.16.100.0"
    search_domain   = "example.com"

    permissions = "642"
    group = "oneadmin"
    security_groups = [0]
    tags = {
        env = "dev"
        customer = "test"
    }
}

resource "opennebula_virtual_network_address_range" "test_ar_1" {
    virtual_network_id = opennebula_virtual_network.test_vnet_address_ranges.id
    ar_type            = "IP4"
    size               = 16
    ip4                = "172.16.100.110"
    hold_ips           = ["172.16.100.112"]
    mac                = "02:00:00:00:00:01"
    custom             = {
        key1 = "value1"
        key2 = "value2"
    }
}

// create the rest of resources for passing the test
resource "opennebula_virtual_network_address_range" "test_ar_2" {
	virtual_network_id = opennebula_virtual_network.test_vnet_address_ranges.id
	ar_type            = "IP6_STATIC"
    size               = 15
    ip6                = "2001:db8::1"
    prefix_length      = "64"
    mac                = "02:00:00:00:01:01"
}

resource "opennebula_virtual_network_address_range" "test_ar_3" {
	virtual_network_id = opennebula_virtual_network.test_vnet_address_ranges.id
	ar_type            = "IP4"
	size               = 32
	ip4                = "172.16.150.100"
}

resource "opennebula_virtual_network_address_range" "test_ar_4" {
	virtual_network_id = opennebula_virtual_network.test_vnet_address_ranges.id
	ar_type            = "IP4"
	size               = 3
	ip4                = "192.168.0.200"
}
`

var testAccDataSourceVirtualNetworkAddressRanges = `
data "opennebula_virtual_network_address_ranges" "test_ar" {
	virtual_network_id = opennebula_virtual_network.test_vnet_address_ranges.id
}
`
