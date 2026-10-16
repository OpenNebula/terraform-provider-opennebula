package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func TestAccVirtualMachineAutostart(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineNoAutostartConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-vm-autostart"),
					resource.TestCheckNoResourceAttr("opennebula_virtual_machine.test", "autostart"),
					testAccCheckVirtualMachineNoAutostartValue("opennebula_virtual_machine.test"),
				),
			},
			{
				Config: testAccVirtualMachineAutostartConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-vm-autostart"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "autostart", "yes"),
					testAccCheckVirtualMachineAutostartValue("opennebula_virtual_machine.test", "yes"),
				),
			},
			{
				Config: testAccVirtualMachineAutostartUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-vm-autostart"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "autostart", "no"),
					testAccCheckVirtualMachineAutostartValue("opennebula_virtual_machine.test", "no"),
				),
			},
		},
	})
}

func testAccCheckVirtualMachineAutostartValue(n string, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		controller := testAccProvider.Meta().(*goca.Controller)
		vmC := controller.VM(toInt(rs.Primary.ID))
		vm, err := vmC.Info(false)
		if err != nil {
			return err
		}

		got, err := vm.UserTemplate.GetStr("AUTOSTART")
		if err != nil {
			return fmt.Errorf("failed to get AUTOSTART from USER_TEMPLATE: %s", err)
		}
		if got != expected {
			return fmt.Errorf("wrong autostart value in USER_TEMPLATE, got %s instead of %s", got, expected)
		}

		return nil
	}
}

func testAccCheckVirtualMachineNoAutostartValue(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		controller := testAccProvider.Meta().(*goca.Controller)
		vmC := controller.VM(toInt(rs.Primary.ID))
		vm, err := vmC.Info(false)
		if err != nil {
			return err
		}

		// Check if AUTOSTART exists in USER_TEMPLATE
		if _, err := vm.UserTemplate.GetStr("AUTOSTART"); err == nil {
			return fmt.Errorf("AUTOSTART element should not exist in USER_TEMPLATE when autostart is not defined")
		}

		return nil
	}
}

var testAccVirtualMachineNoAutostartConfig = `
resource "opennebula_virtual_machine" "test" {
	name        = "test-vm-autostart"
	description = "VM without autostart setting"
	cpu         = 1
	vcpu        = 1
	memory      = 128

	context = {
		NETWORK  = "YES"
		SET_HOSTNAME = "$NAME"
	}
}
`

var testAccVirtualMachineAutostartConfig = `
resource "opennebula_virtual_machine" "test" {
	name        = "test-vm-autostart"
	description = "VM with autostart enabled"
	cpu         = 1
	vcpu        = 1
	memory      = 128
	autostart   = "yes"

	context = {
		NETWORK  = "YES"
		SET_HOSTNAME = "$NAME"
	}
}
`

var testAccVirtualMachineAutostartUpdateConfig = `
resource "opennebula_virtual_machine" "test" {
	name        = "test-vm-autostart"
	description = "VM with autostart disabled"
	cpu         = 1
	vcpu        = 1
	memory      = 128
	autostart   = "no"

	context = {
		NETWORK  = "YES"
		SET_HOSTNAME = "$NAME"
	}
}
`