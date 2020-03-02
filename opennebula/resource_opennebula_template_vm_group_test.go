package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"reflect"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func TestAccVMGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVMGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMGroupConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "name", "test-vmgroup"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "role.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "role.0.name", "anti-aff"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "role.0.policy", "ANTI_AFFINED"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "role.1.name", "host-aff"),
					testAccCheckVMGroupPermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
				),
			},
			{
				Config: testAccVMGroupConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "name", "test-vmgroup-up"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "tags.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "tags.version", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "role.0.name", "anti-aff"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "role.0.policy", "ANTI_AFFINED"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine_group.test", "role.1.name", "host-aff"),
					testAccCheckVMGroupPermissions(&shared.Permissions{
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

func testAccCheckVMGroupDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opennebula_virtual_machine_group" {
			continue
		}
		vmgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		vmgc := controller.VMGroup(int(vmgID))
		// Get Virtual Machine Group Info
		vmg, _ := vmgc.Info(false)
		if vmg != nil {
			return fmt.Errorf("Expected VM Group %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckVMGroupPermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "opennebula_virtual_machine_group" {
				continue
			}
			vmgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vmgc := controller.VMGroup(int(vmgID))
			// Get Virtual Machine Group Info
			vmg, _ := vmgc.Info(false)
			if vmg == nil {
				return fmt.Errorf("Expected VM group %s to exist when checking permissions", rs.Primary.ID)
			}

			if !reflect.DeepEqual(vmg.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for VM Group %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(*expected),
					permissionsUnixString(*vmg.Permissions),
				)
			}
		}

		return nil
	}
}

var testAccVMGroupConfigBasic = `
resource "opennebula_virtual_machine_group" "test" {
  name        = "test-vmgroup"
  group       = "oneadmin"
  permissions = "642"
  role {
    name = "anti-aff"
    policy = "ANTI_AFFINED"
  }
  role {
    name = "host-aff"
    host_affined = [ 0 ]
  }
  tags = {
    env = "prod"
    customer = "test"
  }
}
`

var testAccVMGroupConfigUpdate = `
resource "opennebula_virtual_machine_group" "test" {
  name        = "test-vmgroup-up"
  group       = "oneadmin"
  permissions = "660"
  role {
    name = "anti-aff"
    policy = "ANTI_AFFINED"
  }
  role {
    name = "host-aff"
    host_affined = [ 0 ]
  }
  tags = {
    env = "dev"
    customer = "test"
    version = "2"
  }
}
`
