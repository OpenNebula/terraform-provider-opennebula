package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVirtualRouter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualRouterConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					// Virtual router instance template
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance_template.test", "name", "testacc-vr-template"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance_template.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance_template.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance_template.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance_template.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance_template.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance_template.test", "gname"),
					// Vritual router instance
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test", "name", "testacc-vr-virtual-machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "gname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "virtual_router_id"),
					// Virtual router
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "name", "testacc-vr"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gname"),
					testAccCheckVirtualRouterPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}, "testacc-vr"),
				),
			},
			{
				Config: testAccVirtualRouterAddMachine,
				Check: resource.ComposeTestCheckFunc(
					// First virtual machine
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test", "name", "testacc-vr-virtual-machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "gname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test", "virtual_router_id"),
					// Second virtual machine
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test2", "name", "testacc-vr-virtual-machine-2"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test2", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test2", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test2", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test2", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test2", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test2", "gname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_instance.test2", "virtual_router_id"),
					// Virtual router
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "name", "testacc-vr"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gname"),
					testAccCheckVirtualRouterPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}, "testacc-vr"),
				),
			},
			{
				Config: testAccVirtualRouterContextUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_router_instance.test2", "context.update_test", "123"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gname"),
					testAccCheckVirtualRouterPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}, "testacc-vr"),
				),
			},
			{
				Config: testAccVirtualRouterAddNICs,
				Check: resource.ComposeTestCheckFunc(
					// Virtual router
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "name", "testacc-vr"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_nic.nic1", "network_id"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_nic.nic2", "network_id"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic1", "floating_ip", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic1", "floating_only", "true"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_nic.nic1", "ip"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic2", "floating_ip", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic2", "floating_only", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic2", "ip", ""),
					testAccCheckVirtualRouterPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}, "testacc-vr"),
				),
			},
			{
				Config: testAccVirtualRouterUpdateNICs,
				Check: resource.ComposeTestCheckFunc(
					// Virtual router
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "name", "testacc-vr"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_nic.nic1", "network_id"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router_nic.nic2", "network_id"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic1", "floating_ip", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic1", "floating_only", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic2", "floating_ip", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic2", "floating_only", "false"),
					testAccCheckVirtualRouterPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}, "testacc-vr"),
				),
			},
			{
				Config: testAccVirtualRouterDeleteSecondVrouter,
				Check: resource.ComposeTestCheckFunc(
					// Virtual router
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "name", "testacc-vr"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gname"),
					testAccCheckVirtualRouterPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}, "testacc-vr"),
				),
			},
			{
				Config: testAccVirtualRouterAddNICsWithIPs,
				Check: resource.ComposeTestCheckFunc(
					// Virtual router
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "name", "testacc-vr"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gname"),
					// IP4
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_ip_specified", "floating_ip", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_ip_specified", "floating_only", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_ip_specified", "ip", "172.16.100.160"),
					// IP6
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip6_specified", "floating_ip", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip6_specified", "floating_only", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip6_specified", "ip6", "fd00:ffff:ffff::1"),
					testAccCheckVirtualRouterPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}, "testacc-vr"),
				),
			},
			{
				Config: testAccVirtualRouterUpdateNICsWithIPs,
				Check: resource.ComposeTestCheckFunc(
					// Virtual router
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "name", "testacc-vr"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_router.test", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_router.test", "gname"),
					// IP4
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip_specified", "floating_ip", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip_specified", "floating_only", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip_specified", "ip", "172.16.100.120"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_only_ip_specified", "floating_ip", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_only_ip_specified", "floating_only", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_only_ip_specified", "ip", "172.16.100.121"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_ip_specified", "floating_ip", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_ip_specified", "floating_only", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_floating_ip_specified", "ip", "172.16.100.160"),
					// IP6
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip6_specified", "floating_ip", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip6_specified", "floating_only", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_router_nic.nic_ip6_specified", "ip6", "fd00:ffff:ffff::1"),
					testAccCheckVirtualRouterPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}, "testacc-vr"),
				),
			},
		},
	})
}

func testAccCheckVirtualRouterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "opennebula_virtual_router":
			vrID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			vrc := controller.VirtualRouter(int(vrID))
			vr, _ := vrc.Info(false)
			if vr != nil {
				return fmt.Errorf("Expected virtual router %s to have been destroyed", rs.Primary.ID)
			}
		case "opennebula_virtual_network":
			vnID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			vnc := controller.Template(int(vnID))
			// Get Virtual Machine Info
			vn, _ := vnc.Info(false, false)
			if vn != nil {
				return fmt.Errorf("Expected virtual network %s to have been destroyed", rs.Primary.ID)
			}
		case "opennebula_virtual_router_instance":
			vmID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			vmc := controller.VM(int(vmID))
			// Get Virtual Machine Info
			vm, _ := vmc.Info(false)
			if vm != nil {
				vmState, _, _ := vm.State()
				if vmState != 6 {
					return fmt.Errorf("Expected virtual machine %s to have been destroyed. vmState: %v", rs.Primary.ID, vmState)
				}
			}
		case "opennebula_virtual_router_instance_template":
			tplID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			tc := controller.Template(int(tplID))
			// Get Virtual Machine Info
			tpl, _ := tc.Info(false, false)
			if tpl != nil {
				return fmt.Errorf("Expected template %s to have been destroyed", rs.Primary.ID)
			}
		default:
		}
	}

	return nil
}

func testAccCheckVirtualRouterPermissions(expected *shared.Permissions, resourcename string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			switch rs.Type {
			case "opennebula_virtual_router":

				vrID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
				vrc := controller.VirtualRouter(int(vrID))
				// Get virtual router Info
				vrInfos, _ := vrc.Info(false)
				if vrInfos == nil {
					return fmt.Errorf("Expected virtual router %s to exist when checking permissions", rs.Primary.ID)
				}
				if vrInfos.Name != resourcename {
					continue
				}

				if !reflect.DeepEqual(vrInfos.Permissions, expected) {
					return fmt.Errorf(
						"Permissions for virtual router %s were expected to be %s. Instead, they were %s",
						rs.Primary.ID,
						permissionsUnixString(*expected),
						permissionsUnixString(*vrInfos.Permissions),
					)
				}
			default:
			}
		}

		return nil
	}
}

var testAccVirtualRouterMachineTemplate = `

resource "opennebula_virtual_router_instance_template" "test" {
	name        = "testacc-vr-template"
	permissions = "642"
	group       = "oneadmin"
	cpu         = "0.5"
	vcpu        = "1"
	memory      = "512"

	context = {
	  dns_hostname = "yes"
	  network = "YES"
	}

	graphics {
	  keymap = "en-us"
	  listen = "0.0.0.0"
	  type = "VNC"
	}

	os {
	  arch = "x86_64"
	  boot = ""
	}

	tags = {
	  env = "prod"
	}
}
`

var testAccVirtualRouterVNet = `

resource "opennebula_virtual_network" "network1" {
	name = "test-net1"
	type            = "bridge"
	bridge          = "onebr"
	mtu             = 1500
	ar {
	  ar_type = "IP4"
	  size    = 12
	  ip4     = "172.16.100.130"
	}
	permissions = "642"
	group = "oneadmin"
	security_groups = [0]
	cluster_ids     = [0]
}

resource "opennebula_virtual_network" "network2" {
	name = "test-net2"
	type            = "bridge"
	bridge          = "onebr"
	mtu             = 1500
	ar {
	  ar_type = "IP4"
	  size    = 16
	  ip4     = "172.16.100.110"
	}
	permissions = "642"
	group = "oneadmin"
	security_groups = [0]
	cluster_ids     = [0]
}

resource "opennebula_virtual_network" "network3" {
	name = "test-net3"
	type            = "bridge"
	bridge          = "onebr"
	mtu             = 1500
	ar {
	  ar_type = "IP4"
	  size    = 16
	  ip4     = "172.16.100.150"
	}
	permissions = "642"
	group = "oneadmin"
	security_groups = [0]
	cluster_ids     = [0]
}
resource "opennebula_virtual_network_address_range" "network3_static_ip6" {
	virtual_network_id = opennebula_virtual_network.network3.id
	ar_type            = "IP6_STATIC"
    ip6     		   = "fd00:ffff:ffff::"
	prefix_length	   = 126
	size			   = 4
}
`

var testAccVirtualRouterConfigBasic = testAccVirtualRouterMachineTemplate + `

resource "opennebula_virtual_router_instance" "test" {
	name        = "testacc-vr-virtual-machine"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}

resource "opennebula_virtual_router" "test" {
    name        = "testacc-vr"
    permissions = "642"
    group       = "oneadmin"

    instance_template_id = opennebula_virtual_router_instance_template.test.id

    tags = {
        customer = "1"
    }
}
`

var testAccVirtualRouterAddMachine = testAccVirtualRouterMachineTemplate + `

resource "opennebula_virtual_router_instance" "test" {
	name        = "testacc-vr-virtual-machine"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}


resource "opennebula_virtual_router_instance" "test2" {
	name        = "testacc-vr-virtual-machine-2"
	group       = "oneadmin"
	permissions = "642"
	memory = 128
	cpu = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}

resource "opennebula_virtual_router" "test" {
    name        = "testacc-vr"
    permissions = "642"
    group       = "oneadmin"

    instance_template_id = opennebula_virtual_router_instance_template.test.id

    tags = {
        customer = "1"
    }
}
`

var testAccVirtualRouterContextUpdate = testAccVirtualRouterMachineTemplate + `

resource "opennebula_virtual_router_instance" "test" {
	name        = "testacc-vr-virtual-machine"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}


resource "opennebula_virtual_router_instance" "test2" {
	name        = "testacc-vr-virtual-machine-2"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id

	context = {
	  update_test = "123"
	}
}

resource "opennebula_virtual_router" "test" {
    name        = "testacc-vr"
    permissions = "642"
    group       = "oneadmin"

    instance_template_id = opennebula_virtual_router_instance_template.test.id

    tags = {
        customer = "1"
    }
}
`

var testAccVirtualRouterAddNICs = testAccVirtualRouterMachineTemplate + testAccVirtualRouterVNet + `

resource "opennebula_virtual_router_instance" "test" {
	name        = "testacc-vr-virtual-machine"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}


resource "opennebula_virtual_router_instance" "test2" {
	name        = "testacc-vr-virtual-machine-2"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}

resource "opennebula_virtual_router" "test" {
    name        = "testacc-vr"
    permissions = "642"
    group       = "oneadmin"

    instance_template_id = opennebula_virtual_router_instance_template.test.id

    tags = {
        customer = "1"
    }
}

resource "opennebula_virtual_router_nic" "nic2" {
    virtual_router_id = opennebula_virtual_router.test.id
    network_id        = opennebula_virtual_network.network2.id
}

resource "opennebula_virtual_router_nic" "nic1" {
    floating_ip       = true
    floating_only     = true
    virtual_router_id = opennebula_virtual_router.test.id
    network_id        = opennebula_virtual_network.network1.id
}
`
var testAccVirtualRouterUpdateNICs = testAccVirtualRouterMachineTemplate + testAccVirtualRouterVNet + `

resource "opennebula_virtual_router_instance" "test" {
	name        = "testacc-vr-virtual-machine"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}


resource "opennebula_virtual_router_instance" "test2" {
	name        = "testacc-vr-virtual-machine-2"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}

resource "opennebula_virtual_router" "test" {
    name        = "testacc-vr"
    permissions = "642"
    group       = "oneadmin"

    instance_template_id = opennebula_virtual_router_instance_template.test.id

    tags = {
        customer = "1"
    }
}

resource "opennebula_virtual_router_nic" "nic2" {
	virtual_router_id = opennebula_virtual_router.test.id
	network_id        = opennebula_virtual_network.network3.id
}

resource "opennebula_virtual_router_nic" "nic1" {
	floating_ip       = false
	virtual_router_id = opennebula_virtual_router.test.id
	network_id        = opennebula_virtual_network.network1.id
}
`

var testAccVirtualRouterDeleteSecondVrouter = testAccVirtualRouterMachineTemplate + testAccVirtualRouterVNet + `

resource "opennebula_virtual_router_instance" "test" {
	name        = "testacc-vr-virtual-machine"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}

resource "opennebula_virtual_router" "test" {
    name        = "testacc-vr"
    permissions = "642"
    group       = "oneadmin"

    instance_template_id = opennebula_virtual_router_instance_template.test.id

    tags = {
        customer = "1"
    }
}
`

var testAccVirtualRouterAddNICsWithIPs = testAccVirtualRouterMachineTemplate + testAccVirtualRouterVNet + `

resource "opennebula_virtual_router_instance" "test" {
	name        = "testacc-vr-virtual-machine"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}

resource "opennebula_virtual_router" "test" {
    name        = "testacc-vr"
    permissions = "642"
    group       = "oneadmin"

    instance_template_id = opennebula_virtual_router_instance_template.test.id

    tags = {
        customer = "1"
    }
}

resource "opennebula_virtual_router_nic" "nic_floating_ip_specified" {
    depends_on        = [opennebula_virtual_router_instance.test]
    ip                = "172.16.100.160"
    floating_ip       = true
    virtual_router_id = opennebula_virtual_router.test.id
    network_id        = opennebula_virtual_network.network3.id
}

resource "opennebula_virtual_router_nic" "nic_ip6_specified" {
    depends_on        = [opennebula_virtual_router_instance.test]
    ip6               = "fd00:ffff:ffff::1"
    virtual_router_id = opennebula_virtual_router.test.id
    network_id        = opennebula_virtual_network.network3.id
}
`

var testAccVirtualRouterUpdateNICsWithIPs = testAccVirtualRouterMachineTemplate + testAccVirtualRouterVNet + `

resource "opennebula_virtual_router_instance" "test" {
	name        = "testacc-vr-virtual-machine"
	group       = "oneadmin"
	permissions = "642"
	memory      = 128
	cpu         = 0.1

	virtual_router_id = opennebula_virtual_router.test.id
}

resource "opennebula_virtual_router" "test" {
    name        = "testacc-vr"
    permissions = "642"
    group       = "oneadmin"

    instance_template_id = opennebula_virtual_router_instance_template.test.id

    tags = {
        customer = "1"
    }
}

resource "opennebula_virtual_router_nic" "nic_ip_specified" {
    depends_on        = [opennebula_virtual_router_instance.test]
    ip                = "172.16.100.120"
    virtual_router_id = opennebula_virtual_router.test.id
    network_id        = opennebula_virtual_network.network2.id
}

resource "opennebula_virtual_router_nic" "nic_ip6_specified" {
    depends_on        = [opennebula_virtual_router_instance.test]
    ip6               = "fd00:ffff:ffff::1"
    virtual_router_id = opennebula_virtual_router.test.id
    network_id        = opennebula_virtual_network.network3.id
}

resource "opennebula_virtual_router_nic" "nic_floating_only_ip_specified" {
    depends_on        = [opennebula_virtual_router_instance.test]
    ip                = "172.16.100.121"
    floating_ip       = true
    floating_only     = true
    virtual_router_id = opennebula_virtual_router.test.id
    network_id        = opennebula_virtual_network.network2.id
}

resource "opennebula_virtual_router_nic" "nic_floating_ip_specified" {
    depends_on        = [opennebula_virtual_router_instance.test]
    ip                = "172.16.100.160"
    floating_ip       = true
    virtual_router_id = opennebula_virtual_router.test.id
    network_id        = opennebula_virtual_network.network3.id
}
`
