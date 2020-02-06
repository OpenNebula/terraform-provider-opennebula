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

func TestAccTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccTemplateConfigBasic,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template", "name", "terra-tpl"),
					resource.TestCheckResourceAttr("opennebula_template.template", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_template.template", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gname"),
					testAccCheckTemplatePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						GroupM: 1,
					}),
				),
			},
			{
				Config:             testAccTemplateConfigUpdate,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template", "name", "terratplupdate"),
					resource.TestCheckResourceAttr("opennebula_template.template", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_template.template", "group", "oneadmin"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gname"),
					testAccCheckTemplatePermissions(&shared.Permissions{
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

func testAccCheckTemplatePermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "opennebula_template" {
				continue
			}
			tID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			tc := controller.Template(int(tID))
			// TODO: fix it after 5.10 release availability
			// Force the "extended" bool to false to keep ONE 5.8 behavior
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			template, _ := tc.Info(false, false)
			if template == nil {
				return fmt.Errorf("Expected template %s to exist when checking permissions", rs.Primary.ID)
			}

			if !reflect.DeepEqual(template.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for template %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(*expected),
					permissionsUnixString(*template.Permissions),
				)
			}
		}

		return nil
	}
}
func testAccCheckTemplateDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)
	var destroy bool

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opennebula_template" {
			continue
		}
		templateID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		tc := controller.Template(int(templateID))
		// Get Template Info
		template, _ := tc.Info(false, false)
		if template != nil {
			return fmt.Errorf("Expected template %s to have been destroyed", rs.Primary.ID)
		}
		destroy = true
	}

	if !destroy {
		return fmt.Errorf("No resource to be destroyed")
	}

	return nil
}

var testAccTemplateConfigBasic = `
resource "opennebula_template" "template" {
  name = "terra-tpl"
  permissions = "660"
  group = "oneadmin"
  template = <<-EOT
    CPU = "0.5"
    VCPU = "1"
    MEMORY = "512"
    CONTEXT = [
      DNS_HOSTNAME = "yes",
      NETWORK = "YES"
    ]
    DISK = []
    GRAPHICS = [
      KEYMAP = "en-us",
      LISTEN = "0.0.0.0",
      TYPE = "VNC"
    ]
    OS = [
      ARCH = "x86_64",
      BOOT = "" ]
    EOT
}
`

var testAccTemplateConfigUpdate = `
resource "opennebula_template" "template" {
  name = "terratplupdate"
  permissions = "642"
  group = "oneadmin"
  template = <<-EOT
    CPU = "1"
    VCPU = "1"
    MEMORY = "768"
    CONTEXT = [
      DNS_HOSTNAME = "yes",
      NETWORK = "YES"
    ]
    DISK = []
    GRAPHICS = [
      KEYMAP = "en-us",
      LISTEN = "0.0.0.0",
      TYPE = "VNC"
    ]
    OS = [
      ARCH = "x86_64",
      BOOT = "" ]
    EOT
}
`
