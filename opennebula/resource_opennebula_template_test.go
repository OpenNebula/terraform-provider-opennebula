package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

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
				Config: testAccTemplateConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template", "name", "terra-tpl"),
					resource.TestCheckResourceAttr("opennebula_template.template", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_template.template", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_template.template", "cpu", "0.5"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_template.template", "labels", "test1,test2,test3"),
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
				Config: testAccTemplateConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template", "name", "terratplupdate"),
					resource.TestCheckResourceAttr("opennebula_template.template", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_template.template", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_template.template", "cpu", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.%", "3"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.version", "2"),
					resource.TestCheckResourceAttr("opennebula_template.template", "labels", "test1,test2,test3,test4"),
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
  cpu = "0.5"
  vcpu = "1"
  memory = "512"

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

  labels = "test1,test2,test3"

  tags = {
    env = "prod"
    customer = "test"
  }
}
`

var testAccTemplateConfigUpdate = `
resource "opennebula_template" "template" {
  name = "terratplupdate"
  permissions = "642"
  group = "oneadmin"

  cpu = "1"
  vcpu = "1"
  memory = "768"

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

  labels = "test1,test2,test3,test4"

  tags = {
    env = "dev"
    customer = "test"
    version = "2"
  }
}
`
