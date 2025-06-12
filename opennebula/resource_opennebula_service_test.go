//go:build !legacy

package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func TestAccService(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_service.test", "name", "service-test-tf-basic"),
					resource.TestCheckResourceAttr("opennebula_service.test", "permissions", "642"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "gname"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "state"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "template_id"),
				),
			},
			{
				Config: testAccServiceConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_service.test", "name", "service-test-tf-renamed"),
					resource.TestCheckResourceAttr("opennebula_service.test", "permissions", "777"),
					resource.TestCheckResourceAttr("opennebula_service.test", "uid", "1"),
					resource.TestCheckResourceAttr("opennebula_service.test", "gid", "1"),
					resource.TestCheckResourceAttr("opennebula_service.test", "uname", "serveradmin"),
					resource.TestCheckResourceAttr("opennebula_service.test", "gname", "users"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "state"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "template_id"),
				),
			},
		},
	})
}

func testAccCheckServiceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opennebula_service" {
			continue
		}
		svID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
		sc := controller.Service(int(svID))
		// Get Service Info
		service, _ := sc.Info()
		if service != nil {
			svState := service.Template.Body.StateRaw
			if svState != 5 {
				return fmt.Errorf("Expected service %s to have been destroyed", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckServicePermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			serviceID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			sc := controller.Service(int(serviceID))
			// Get Service
			service, _ := sc.Info()
			if service == nil {
				return fmt.Errorf("Expected service %s to exist when checking permissions", rs.Primary.ID)
			}

			if !reflect.DeepEqual(service.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for service %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(*expected),
					permissionsUnixString(*service.Permissions),
				)
			}
		}

		return nil
	}
}

var testAccServiceVMTemplate = `

resource "opennebula_template" "test" {
  name = "service-basic-test"

  cpu    = 1
  vcpu   = 1
  memory = 64
}
`

var testAccServiceServiceTemplate = `

resource "opennebula_service_template" "test" {
  name     = "service-basic-template"
 permissions = "644"
  template = jsonencode({
    TEMPLATE = {
      BODY = {
        name       = "service"
        deployment = "straight"
        roles = [
          {
            name        = "master"
            type        = "vm"
            template_id = tonumber(opennebula_template.test.id)
            cardinality = 1
            min_vms     = 1
          }
        ]
      }
    }
  })
}
`

var testAccServiceConfigBasic = testAccServiceVMTemplate + testAccServiceServiceTemplate + `

resource "opennebula_service" "test" {
  name        = "service-test-tf-basic"
  template_id = opennebula_service_template.test.id
  permissions = "642"
  uid         = 0
  gid         = 0
}
`

var testAccServiceConfigUpdate = testAccServiceVMTemplate + testAccServiceServiceTemplate + `

resource "opennebula_service" "test" {
  name        = "service-test-tf-renamed"
  template_id = opennebula_service_template.test.id
  permissions = "777"
  uid         = 1
  gid         = 1
}
`
