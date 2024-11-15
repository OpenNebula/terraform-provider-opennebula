package opennebula

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceOpennebulaVirtualNetworkAddressRange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualNetworkAddressRangeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVirtualNetworkAddressRangeConfig + testAccDataSourceVirtualNetworkAddressRange1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.opennebula_virtual_network_address_range.test_ar_1", "virtual_network_id",
						"opennebula_virtual_network.test_vnet", "id",
					),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ar_type", "IP4"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "size", "16"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ip4", "172.16.100.110"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ip4_end", "172.16.100.125"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "global_prefix", ""),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ula_prefix", ""),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ip6", ""),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ip6_end", ""),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ip6_global", ""),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ip6_global_end", ""),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ip6_ula", ""),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "ip6_ula_end", ""),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "mac", "02:00:00:00:00:01"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "mac_end", "02:00:00:00:00:10"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "held_ips.#", "1"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "held_ips.0", "172.16.100.112"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "custom.%", "2"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "custom.key1", "value1"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_1", "custom.key2", "value2"),
				),
			},
			{
				Config: testAccDataSourceVirtualNetworkAddressRangeConfig + testAccDataSourceVirtualNetworkAddressRange2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.opennebula_virtual_network_address_range.test_ar_2", "virtual_network_id",
						"opennebula_virtual_network.test_vnet", "id",
					),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_2", "ar_type", "IP6_STATIC"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_2", "size", "15"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_2", "ip6", "2001:db8::1"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_2", "ip6_end", "2001:db8::f"),
					resource.TestCheckResourceAttr("data.opennebula_virtual_network_address_range.test_ar_2", "mac", "02:00:00:00:01:01"),
				),
			},
		},
	})
}

func testAccCheckVirtualNetworkAddressRangeDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "opennebula_virtual_network" && rs.Primary.Attributes["name"] == "test_vnet" {
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

var testAccDataSourceVirtualNetworkAddressRangeConfig = `
resource "opennebula_virtual_network" "test_vnet" {
    name            = "test-vnet"
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
    virtual_network_id = opennebula_virtual_network.test_vnet.id
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

resource "opennebula_virtual_network_address_range" "test_ar_2" {
    virtual_network_id = opennebula_virtual_network.test_vnet.id
    ar_type            = "IP6_STATIC"
    size               = 15
    ip6                = "2001:db8::1"
    prefix_length      = "64"
    mac                = "02:00:00:00:01:01"
}
`

var testAccDataSourceVirtualNetworkAddressRange1 = `
data "opennebula_virtual_network_address_range" "test_ar_1" {
	virtual_network_id = opennebula_virtual_network.test_vnet.id
    id = opennebula_virtual_network_address_range.test_ar_1.id
}
`

var testAccDataSourceVirtualNetworkAddressRange2 = `
data "opennebula_virtual_network_address_range" "test_ar_2" {
	virtual_network_id = opennebula_virtual_network.test_vnet.id
    id = opennebula_virtual_network_address_range.test_ar_2.id
}
`
