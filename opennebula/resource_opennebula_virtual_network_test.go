package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"reflect"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	goca "github.com/OpenNebula/one/src/oca/go/src/goca"
)

func TestAccVirtualNetwork(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNetworkConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "name", "test-virtual_network"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "physical_device", "dummy0"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "type", "vxlan"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "vlan_id", "8000046"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "mtu", "1500"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gname"),
					testAccCheckVirtualNetworkARnumber(2),
					testAccCheckVirtualNetworkPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
				),
			},
			{
				Config: testAccVirtualNetworkConfigUpdate,
                ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "name", "test-virtual_network"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "physical_device", "dummy0"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "type", "vxlan"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "vlan_id", "8000046"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "mtu", "1500"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "group", "users"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gname"),
					testAccCheckVirtualNetworkARnumber(2),
					testAccCheckVirtualNetworkPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						OwnerA: 0,
						GroupU: 1,
						GroupM: 1,
					}),
				),
			},
		},
	})
}

func testAccCheckVirtualNetworkDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		vnc := controller.VirtualNetwork(int(vnID))
		// Get Virtual Network Info
		vn, _ := vnc.Info()
		if vn != nil {
			return fmt.Errorf("Expected virtual network %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckVirtualNetworkARnumber(expectedARs int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vnc := controller.VirtualNetwork(int(vnID))
			// Get Virtual Network Info
			vn, _ := vnc.Info()
			if vn == nil {
				return fmt.Errorf("Expected virtual network %s to exist", rs.Primary.ID)
			}

			if len(vn.ARs) != expectedARs {
				return fmt.Errorf("Expected ARs number: %d, got: %d", expectedARs, len(vn.ARs))
			}
		}

		return nil
	}
}

func testAccCheckVirtualNetworkPermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vnc := controller.VirtualNetwork(int(vnID))
			// Get Virtual Network Info
			vn, _ := vnc.Info()
			if vn == nil {
				return fmt.Errorf("Expected virtual_network %s to exist when checking permissions", rs.Primary.ID)
			}

			if !reflect.DeepEqual(vn.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for virtual_network %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(expected),
					permissionsUnixString(vn.Permissions),
				)
			}
		}

		return nil
	}
}

var testAccVirtualNetworkConfigBasic = `
resource "opennebula_virtual_network" "test" {
  name = "test-virtual_network"
  physical_device = "dummy0"
  type            = "vxlan"
  vlan_id         = "8000046"
  mtu             = 1500
  ar {
    ar_type = "IP4"
    size    = 16
    ip4     = "172.16.100.110"
  }
  ar {
    ar_type = "IP4"
    size    = 12
    ip4     = "172.16.100.130"
  }
  permissions = "642"
  group = "oneadmin"
  security_groups = [0]
}
`

var testAccVirtualNetworkConfigUpdate = `
resource "opennebula_virtual_network" "test" {
  name = "test-virtual_network"
  physical_device = "dummy0"
  type            = "vxlan"
  vlan_id         = "8000046"
  mtu             = 1500
  ar {
    ar_type = "IP4"
    size    = 16
    ip4     = "172.16.100.110"
  }
  ar {
    ar_type = "IP4"
    size    = 12
    ip4     = "172.16.100.130"
  }
  permissions = "660"
  group = "users"
  security_groups = [0]
}
`
