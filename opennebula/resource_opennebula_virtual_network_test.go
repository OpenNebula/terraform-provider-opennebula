package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
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
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "bridge", "onebr"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "type", "dummy"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "mtu", "1500"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "dns", "172.16.100.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "gateway", "172.16.100.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "network_mask", "255.255.255.0"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gname"),
					testAccCheckVirtualNetworkARnumber(2),
					testAccVirtualNetworkAR(0, "ar_type", "IP4"),
					testAccVirtualNetworkAR(0, "size", "16"),
					testAccVirtualNetworkAR(0, "ip4", "172.16.100.110"),
					testAccVirtualNetworkAR(1, "ar_type", "IP4"),
					testAccVirtualNetworkAR(1, "size", "12"),
					testAccVirtualNetworkAR(1, "ip4", "172.16.100.130"),
					testAccVirtualNetworkSG([]int{0}),
					testAccCheckVirtualNetworkPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
				),
			},
			{
				Config:             testAccVirtualNetworkConfigUpdate,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "name", "test-virtual_network-renamed"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "bridge", "onebr"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "type", "dummy"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "mtu", "1500"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "group", "users"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "dns", "172.16.100.254"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "gateway", "172.16.100.254"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "network_mask", "255.255.0.0"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gname"),
					testAccVirtualNetworkAR(0, "ar_type", "IP4"),
					testAccVirtualNetworkAR(0, "size", "16"),
					testAccVirtualNetworkAR(0, "ip4", "172.16.100.110"),
					testAccVirtualNetworkAR(0, "mac", "02:01:ac:10:64:6e"),
					testAccVirtualNetworkAR(1, "ar_type", "IP4"),
					testAccVirtualNetworkAR(1, "size", "13"),
					testAccVirtualNetworkAR(1, "ip4", "172.16.100.130"),
					testAccVirtualNetworkAR(2, "ar_type", "IP6"),
					testAccVirtualNetworkAR(2, "size", "2"),
					testAccVirtualNetworkAR(2, "ip6", "2001:db8:0:85a3::ac1f:8001"),
					testAccCheckVirtualNetworkARnumber(3),
					testAccVirtualNetworkSG([]int{0}),
					testAccCheckVirtualNetworkPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						OwnerA: 0,
						GroupU: 1,
						GroupM: 1,
					}),
				),
			},
			{
				Config:             testAccVirtualNetworkReservationConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_network.reservation", "name", "terravnetres"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.reservation", "reservation_size", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.reservation", "permissions", "660"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.reservation", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.reservation", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.reservation", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.reservation", "gname"),
					testAccCheckVirtualNetworkPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
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
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		vn, _ := vnc.Info(false)
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
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			vn, _ := vnc.Info(false)
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
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			vn, _ := vnc.Info(false)
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

func testAccVirtualNetworkAR(aridx int, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vnc := controller.VirtualNetwork(int(vnID))
			// Get Virtual Network Info
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			vn, _ := vnc.Info(false)
			if vn == nil {
				return fmt.Errorf("Expected virtual network %s to exist when checking permissions", rs.Primary.ID)
			}
			ars := generateARMapFromStructs(vn.ARs)

			var found bool

			for i, ar := range ars {
				if i == aridx {
					if ar[key] != nil && ar[key].(string) != value {
						return fmt.Errorf("Expected %s = %s for AR ID %d, got %s = %s", key, value, aridx, key, ar[key].(string))
					}
					found = true
				}
			}

			if !found {
				return fmt.Errorf("AR id %d with %s = %s does not exist", aridx, key, value)
			}
		}
		return nil
	}
}

func testAccVirtualNetworkSG(slice []int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vnc := controller.VirtualNetwork(int(vnID))
			// Get Virtual Network Info
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			vn, _ := vnc.Info(false)
			if vn == nil {
				return fmt.Errorf("Expected virtual network %s to exist when checking permissions", rs.Primary.ID)
			}
			secgrouplist, err := vn.Template.Dynamic.GetContentByName("SECURITY_GROUPS")
			if err != nil {
				return err
			}
			secgroups_str := strings.Split(secgrouplist, ",")
			secgroups_int := []int{}

			for _, i := range secgroups_str {
				if i != "" {
					j, err := strconv.Atoi(i)
					if err != nil {
						return err
					}
					secgroups_int = append(secgroups_int, j)
				}
			}
			if !reflect.DeepEqual(secgroups_int, slice) {
				return fmt.Errorf("Securty Groups for Virtual Network %s are not the expected ones", rs.Primary.ID)
			}
		}
		return nil
	}
}

var testAccVirtualNetworkConfigBasic = `
resource "opennebula_virtual_network" "test" {
  name = "test-virtual_network"
  type            = "dummy"
  bridge          = "onebr"
  mtu             = 1500
  gateway         = "172.16.100.1"
  dns             = "172.16.100.1"
  network_mask    = "255.255.255.0"
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
  clusters = [0]
}
`

var testAccVirtualNetworkConfigUpdate = `
resource "opennebula_virtual_network" "test" {
  name = "test-virtual_network-renamed"
  type            = "dummy"
  bridge          = "onebr"
  mtu             = 1500
  gateway         = "172.16.100.254"
  dns             = "172.16.100.254"
  network_mask    = "255.255.0.0"
  ar {
    ar_type = "IP4"
    size    = 16
    mac     = "02:01:ac:10:64:6e"
    ip4     = "172.16.100.110"
  }
  ar {
    ar_type = "IP4"
    size    = 13
    ip4     = "172.16.100.140"
  }
  ar {
    ar_type = "IP6"
    size    = 2
    ip6     = "2001:db8:0:85a3::ac1f:8001"
  }
  security_groups = [0]
  clusters = [0]
  permissions = "660"
  group = "users"
}
`

var testAccVirtualNetworkReservationConfig = `
resource "opennebula_virtual_network" "test" {
  name = "test-virtual_network-renamed"
  type            = "dummy"
  bridge          = "onebr"
  mtu             = 1500
  gateway         = "172.16.100.254"
  dns             = "172.16.100.254"
  network_mask    = "255.255.0.0"
  ar {
    ar_type = "IP4"
    size    = 16
    mac     = "02:01:ac:10:64:6e"
    ip4     = "172.16.100.110"
  }
  ar {
    ar_type = "IP4"
    size    = 13
    ip4     = "172.16.100.140"
  }
  ar {
    ar_type = "IP6"
    size    = 2
    ip6     = "2001:db8:0:85a3::ac1f:8001"
  }
  security_groups = [0]
  clusters = [0]
  permissions = "660"
  group = "users"
}

resource "opennebula_virtual_network" "reservation" {
    name = "terravnetres"
    description = "my terraform vnet"
    reservation_vnet = "${opennebula_virtual_network.test.id}"
    reservation_size = 1
    security_groups = [0]
    permissions = 660
}
`
