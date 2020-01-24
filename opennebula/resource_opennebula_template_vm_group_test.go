package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"reflect"
	"strconv"
	"strings"
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
					testAccVMGRole(0, "name", "anti-aff"),
					testAccVMGRole(0, "policy", "ANTI_AFFINED"),
					testAccVMGRole(1, "name", "host-aff"),
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
					testAccVMGRole(0, "name", "anti-aff"),
					testAccVMGRole(0, "policy", "ANTI_AFFINED"),
					testAccVMGRole(1, "name", "host-aff"),
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

func testAccVMGRole(roleidx int, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			vmgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vmgc := controller.VMGroup(int(vmgID))
			vmg, _ := vmgc.Info(false)
			if vmg == nil {
				return fmt.Errorf("Expected VM Group %s to exist when checking permissions", rs.Primary.ID)
			}
			var roles []map[string]interface{}

			for _, vmgr := range vmg.Roles {

				hostAffString := strings.Split(vmgr.HostAffined, ",")
				hostAntiAffString := strings.Split(vmgr.HostAntiAffined, ",")
				vmsString := strings.Split(vmgr.VMs, ",")
				hAff := make([]int, 0)
				hAntiAff := make([]int, 0)
				vms := make([]int, 0)
				for _, h := range hostAffString {
					hostAffInt, _ := strconv.ParseInt(h, 10, 32)
					hAff = append(hAff, int(hostAffInt))
				}
				for _, h := range hostAntiAffString {
					hostAntiAffInt, _ := strconv.ParseInt(h, 10, 32)
					hAntiAff = append(hAff, int(hostAntiAffInt))
				}
				for _, vm := range vmsString {
					vmInt, _ := strconv.ParseInt(vm, 10, 32)
					vms = append(vms, int(vmInt))
				}
				roles = append(roles, map[string]interface{}{
					"id":                vmgr.ID,
					"name":              vmgr.Name,
					"host_affined":      hAff,
					"host_anti_affined": hAntiAff,
					"policy":            vmgr.Policy,
					"vms":               vms,
				})
			}

			var found bool

			for _, role := range roles {
				if roleidx == role["id"] {
					if role[key] != nil && role[key].(string) != value {
						return fmt.Errorf("Expected %s = %s for role ID %d, got %s = %s", key, value, roleidx, key, role[key].(string))
					}
					found = true
				} else {
					continue
				}
			}

			if !found {
				return fmt.Errorf("role id %d with %s = %s does not exist, %v, %v, %v", roleidx, key, value, roles, vmg.Roles, vmg)
			}

		}

		return nil
	}
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
					permissionsUnixString(expected),
					permissionsUnixString(vmg.Permissions),
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
}
`
