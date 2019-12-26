package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	ds "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/datastore"
	dskeys "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/datastore/keys"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func TestAccVirtualMachine(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineTemplateConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "128"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.1"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gname"),
					testAccCheckVirtualMachinePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
				),
			},
			{
				Config: testAccVirtualMachineConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine-renamed"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "196"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.2"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gname"),
					testAccCheckVirtualMachinePermissions(&shared.Permissions{
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

func TestAccVirtualMachinePending(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachinePending,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "virtual_machine_pending"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "128"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "pending", "true"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gname"),
					testAccCheckVirtualMachinePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
				),
			},
		},
	})
}
func testAccCheckVirtualMachineDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opennebula_virtual_machine" {
			continue
		}
		vmID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		vmc := controller.VM(int(vmID))
		// Get Virtual Machine Info
		vm, _ := vmc.Info(false)
		if vm != nil {
			vmState, _, _ := vm.State()
			if vmState != 6 {
				return fmt.Errorf("Expected virtual machine %s to have been destroyed", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccSetDSdummy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v := os.Getenv("TF_ACC_VM"); v == "1" {
			controller := testAccProvider.Meta().(*goca.Controller)

			dstpl := ds.NewTemplate()
			dstpl.Add(dskeys.TMMAD, "dummy")
			dstpl.Add(dskeys.DSMAD, "dummy")
			controller.Datastore(0).Update(dstpl.String(), 1)
			controller.Datastore(1).Update(dstpl.String(), 1)
		}
		return nil
	}
}

func testAccCheckVirtualMachinePermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			vmID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vmc := controller.VM(int(vmID))
			// Get Virtual Machine Info
			vm, _ := vmc.Info(false)
			if vm == nil {
				return fmt.Errorf("Expected virtual_machine %s to exist when checking permissions", rs.Primary.ID)
			}

			if !reflect.DeepEqual(vm.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for virtual_machine %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(expected),
					permissionsUnixString(vm.Permissions),
				)
			}
		}

		return nil
	}
}

var testAccVirtualMachineTemplateConfigBasic = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine"
  group       = "oneadmin"
  permissions = "642"
  memory = 128
  cpu = 0.1

  context = {
    NETWORK  = "YES"
    SET_HOSTNAME = "$NAME"
  }

  graphics {
    type   = "VNC"
    listen = "0.0.0.0"
    keymap = "en-us"
  }

  os {
    arch = "x86_64"
    boot = ""
  }
}
`

var testAccVirtualMachineConfigUpdate = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine-renamed"
  group       = "oneadmin"
  permissions = "660"
  memory = 196
  cpu = 0.2

  context = {
    NETWORK  = "YES"
    SET_HOSTNAME = "$NAME"
  }

  graphics {
    type   = "VNC"
    listen = "0.0.0.0"
    keymap = "en-us"
  }

  os {
    arch = "x86_64"
    boot = ""
  }
}
`

var testAccVirtualMachinePending = `
resource "opennebula_virtual_machine" "test" {
  name        = "virtual_machine_pending"
  group       = "oneadmin"
  permissions = "642"
  memory = 128
  cpu = 0.1
  pending = true

  context = {
    NETWORK  = "YES"
    SET_HOSTNAME = "$NAME"
  }

  graphics {
    type   = "VNC"
    listen = "0.0.0.0"
    keymap = "en-us"
  }

  os {
    arch = "x86_64"
    boot = ""
  }
}
`
