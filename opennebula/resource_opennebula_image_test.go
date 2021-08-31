package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func TestAccImage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfigDatablockBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_image.testimage", "name", "test-image-datablock"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "datastore_id", "1"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "persistent", "true"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "type", "DATABLOCK"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "size", "128"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "computed_size", "128"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "dev_prefix", "vd"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "driver", "qcow2"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "permissions", "742"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "timeout", "5"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "gname"),
					testAccCheckImagePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						OwnerA: 1,
						GroupU: 1,
						OtherM: 1,
					}, "test-image-datablock"),
				),
			},
			{
				Config: testAccImageConfigDatablockLocked,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_image.testimage", "name", "test-image-datablock"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "datastore_id", "1"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "persistent", "true"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "type", "DATABLOCK"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "size", "128"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "computed_size", "128"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "dev_prefix", "vd"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "driver", "qcow2"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "permissions", "742"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "timeout", "5"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "lock", "USE"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "gname"),
					testAccCheckImagePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						OwnerA: 1,
						GroupU: 1,
						OtherM: 1,
					}, "test-image-datablock"),
				),
			},
			{
				Config: testAccImageConfigDatablockUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_image.testimage", "name", "test-image-datablock"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "datastore_id", "1"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "persistent", "false"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "type", "DATABLOCK"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "size", "128"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "computed_size", "128"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "dev_prefix", "vd"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "driver", "qcow2"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.%", "3"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "tags.version", "2"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "timeout", "10"),
					resource.TestCheckResourceAttr("opennebula_image.testimage", "lock", "UNLOCK"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_image.testimage", "gname"),
					testAccCheckImagePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						OwnerA: 0,
						GroupU: 1,
						GroupM: 1,
					}, "test-image-datablock"),
				),
			},
		},
	})
}

func testAccCheckImageDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		imageID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		ic := controller.Image(int(imageID))
		// Get Image Info
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		image, _ := ic.Info(false)
		if image != nil {
			return fmt.Errorf("Expected image %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckImagePermissions(expected *shared.Permissions, resourcename string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			imageID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			ic := controller.Image(int(imageID))
			// Get image Info
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			image, _ := ic.Info(false)
			if image == nil {
				return fmt.Errorf("Expected image %s to exist when checking permissions", rs.Primary.ID)
			}
			if image.Name != resourcename {
				continue
			}

			if !reflect.DeepEqual(image.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for image %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(*expected),
					permissionsUnixString(*image.Permissions),
				)
			}
		}

		return nil
	}
}

var testAccImageConfigDatablockBasic = `
resource "opennebula_image" "testimage" {
   name = "test-image-datablock"
   description = "Terraform datablock"
   datastore_id = 1
   persistent = true
   type = "DATABLOCK"
   size = "128"
   dev_prefix = "vd"
   permissions = "742"
   driver = "qcow2"
   tags = {
     env = "prod"
     customer = "test"
   }
   timeout = 5
}
`

var testAccImageConfigDatablockLocked = `
resource "opennebula_image" "testimage" {
   name = "test-image-datablock"
   description = "Terraform datablock"
   datastore_id = 1
   persistent = true
   type = "DATABLOCK"
   size = "128"
   dev_prefix = "vd"
   permissions = "742"
   driver = "qcow2"
   tags = {
     env = "prod"
     customer = "test"
   }
   timeout = 5
   lock = "USE"
}
`

var testAccImageConfigDatablockUpdate = `
resource "opennebula_image" "testimage" {
   name = "test-image-datablock"
   description = "Terraform datablock"
   datastore_id = 1
   persistent = false
   type = "DATABLOCK"
   size = "128"
   dev_prefix = "vd"
   permissions = 660
   driver = "qcow2"
   tags = {
     env = "dev"
     customer = "test"
     version = "2"
   }
   timeout = 10
   lock = "UNLOCK"
}
`
