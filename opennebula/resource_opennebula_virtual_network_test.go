package opennebula

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func TestAccVirtualNetwork(t *testing.T) {
	networkNotFoundErr, _ := regexp.Compile("Error getting virtual network.*[\n]?.*\\[25\\]")
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
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "network_address", "172.16.100.0"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "search_domain", "example.com"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gname"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "ar_type", "IP4"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "size", "16"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "ip4", "172.16.100.110"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "hold_ips.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "hold_ips.0", "172.16.100.112"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test2", "ar_type", "IP4"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test2", "size", "15"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test2", "ip4", "172.16.100.170"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "ar_type", "IP4"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "size", "12"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "ip4", "172.16.100.130"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "hold_ips.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "hold_ips.0", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "hold_ips.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "hold_ips.0", "172.16.100.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "ar.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_virtual_network.test", "ar.*", map[string]string{
						"ar_type": "IP4",
						"size":    "5",
						"ip4":     "172.16.100.1",
					}),
					testAccVirtualNetworkSG([]int{0}),
					testAccCheckVirtualNetworkPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
					resource.TestCheckTypeSetElemAttr("opennebula_virtual_network.test", "cluster_ids.*", "0"),
				),
			},
			{
				Config: testAccVirtualNetworkConfigUpdate,
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
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "network_address", "172.16.100.0"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "search_domain", "example.com"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_network.test", "gname"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "ar_type", "IP4"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "size", "17"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "ip4", "172.16.100.110"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "hold_ips.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test", "hold_ips.0", "172.16.100.112"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test2", "ar_type", "IP4"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test2", "size", "15"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test2", "ip4", "172.16.100.170"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "ar_type", "IP4"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "size", "13"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "ip4", "172.16.100.140"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "hold_ips.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test3", "hold_ips.0", "172.16.100.141"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test4", "ar_type", "IP6"),
					resource.TestCheckResourceAttr("opennebula_virtual_network_address_range.test4", "size", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "tags.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "tags.version", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "hold_ips.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "hold_ips.0", "172.16.100.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "ar.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_virtual_network.test", "ar.*", map[string]string{
						"ar_type": "IP4",
						"size":    "5",
						"ip4":     "172.16.100.1",
					}),
					testAccVirtualNetworkSG([]int{0}),
					testAccCheckVirtualNetworkPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						OwnerA: 0,
						GroupU: 1,
						GroupM: 1,
					}),
					resource.TestCheckTypeSetElemAttr("opennebula_virtual_network.test", "cluster_ids.*", "0"),
				),
			},
			{
				Config:             testAccVirtualNetworkReservationConfig,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_network.reservation", "name", "terravnetres"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.reservation", "reservation_size", "5"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.reservation", "reservation_first_ip", "172.16.100.115"),
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
					resource.TestCheckTypeSetElemAttr("opennebula_virtual_network.test", "cluster_ids.*", "0"),
				),
			},
			{
				Config:      testAccVirtualNetworkReservationNoNetworkConfig,
				ExpectError: networkNotFoundErr,
			},
		},
	})
}

func testAccCheckVirtualNetworkDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "opennebula_virtual_network":
			vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			vnc := controller.VirtualNetwork(int(vnID))

			// Wait for Virtual Network deleted
			stateConf := &resource.StateChangeConf{
				Pending: []string{"anythingelse"},
				Target:  []string{""},
				Refresh: func() (interface{}, string, error) {

					vn, _ := vnc.Info(false)
					if vn == nil {
						return vn, "", nil
					}

					return vn, "EXISTS", nil
				},
				Timeout:    1 * time.Minute,
				Delay:      10 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			_, err := stateConf.WaitForStateContext(context.Background())

			return err
		default:
		}
	}

	return nil
}

func testAccCheckVirtualNetworkPermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			switch rs.Type {
			case "opennebula_virtual_network":
				vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
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
						permissionsUnixString(*expected),
						permissionsUnixString(*vn.Permissions),
					)
				}
			default:
			}
		}

		return nil
	}
}

func testAccVirtualNetworkSG(slice []int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			switch rs.Type {
			case "opennebula_virtual_network":
				vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
				vnc := controller.VirtualNetwork(int(vnID))
				// Get Virtual Network Info
				// TODO: fix it after 5.10 release
				// Force the "decrypt" bool to false to keep ONE 5.8 behavior
				vn, err := vnc.Info(false)
				if err != nil {
					return fmt.Errorf("Virtual network (ID: %s): failed to retrieve informations: %s", rs.Primary.ID, err)
				}

				secgrouplist, err := vn.Template.Get("SECURITY_GROUPS")
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
			default:
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
  network_address = "172.16.100.0"
  search_domain   = "example.com"
  ar {
    ar_type = "IP4"
    size    = 5
    ip4     = "172.16.100.1"
  }
  hold_ips           = ["172.16.100.2"]

  permissions = "642"
  group = "oneadmin"
  security_groups = [0]
  cluster_ids = [0]
  tags = {
    env = "prod"
    customer = "test"
  }

  lifecycle {
    ignore_changes = [ar, hold_ips, clusters]
  }
}

resource "opennebula_virtual_network_address_range" "test" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 16
	ip4                = "172.16.100.110"
	hold_ips           = ["172.16.100.112"]
}

resource "opennebula_virtual_network_address_range" "test2" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 15
	ip4                = "172.16.100.170"
}

resource "opennebula_virtual_network_address_range" "test3" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 12
	ip4                = "172.16.100.130"
	hold_ips           = ["172.16.100.131"]
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
  network_address = "172.16.100.0"
  search_domain   = "example.com"
  ar {
    ar_type = "IP4"
    size    = 5
    ip4     = "172.16.100.1"
  }
  hold_ips           = ["172.16.100.2"]
  security_groups = [0]
  cluster_ids = [0]
  permissions = "660"
  group = "users"
  tags = {
    env = "dev"
    customer = "test"
    version = "2"
  }

  lifecycle {
    ignore_changes = [ar, hold_ips, clusters]
  }
}

resource "opennebula_virtual_network_address_range" "test" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 17
	ip4                = "172.16.100.110"
	hold_ips = ["172.16.100.112"]
}

resource "opennebula_virtual_network_address_range" "test2" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 15
	ip4                = "172.16.100.170"
}

resource "opennebula_virtual_network_address_range" "test3" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 13
	ip4                = "172.16.100.140"
	hold_ips           = ["172.16.100.141"]
}

resource "opennebula_virtual_network_address_range" "test4" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP6"
	size               = 2
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
    size    = 5
    ip4     = "172.16.100.1"
  }
  hold_ips           = ["172.16.100.2"]
  security_groups = [0]
  cluster_ids = [0]
  permissions = "660"
  group = "users"

  lifecycle {
    ignore_changes = [ar, hold_ips, clusters]
  }
}

resource "opennebula_virtual_network_address_range" "test" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 16
	ip4                = "172.16.100.110"
	hold_ips           = ["172.16.100.112"]
}

resource "opennebula_virtual_network_address_range" "test2" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 15
	ip4                = "172.16.100.170"
}

resource "opennebula_virtual_network_address_range" "test3" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP4"
	size               = 13
	ip4                = "172.16.100.140"
	hold_ips           = ["172.16.100.141"]
}

resource "opennebula_virtual_network_address_range" "test4" {
	virtual_network_id = opennebula_virtual_network.test.id
	ar_type            = "IP6"
	size               = 2
}

resource "opennebula_virtual_network" "reservation" {
    name = "terravnetres"
    description = "my terraform vnet"
    reservation_vnet = opennebula_virtual_network.test.id
    reservation_size = 5
	reservation_ar_id = opennebula_virtual_network_address_range.test.id
	reservation_first_ip = "172.16.100.115"
    security_groups = [0]
    permissions = 660
}
`

var testAccVirtualNetworkReservationNoNetworkConfig = `
resource "opennebula_virtual_network" "non-existing-reservation" {
    name = "terravnetreswqerwer"
    description = "my terraform vnet"
    reservation_vnet = 25
    reservation_size = 1
    security_groups = [0]
    permissions = 660
}
`
