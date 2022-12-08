package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatastore(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatastoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatastoreConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_datastore.example", "name", "test"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_datastore.example", "custom.*", map[string]string{
						"datastore": "dummy",
						"transfer":  "dummy",
					}),
					resource.TestCheckResourceAttr("opennebula_datastore.example", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_datastore.example", "tags.test", "test"),
					resource.TestCheckResourceAttr("opennebula_datastore.example", "tags.environment", "example"),
				),
			},
			{
				Config: testAccDatastoreConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_datastore.example", "name", "test-updated"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_datastore.example", "custom.*", map[string]string{
						"datastore": "dummy",
						"transfer":  "dummy",
					}),
					resource.TestCheckResourceAttr("opennebula_datastore.example", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_datastore.example", "tags.environment", "example-updated"),
					resource.TestCheckResourceAttr("opennebula_datastore.example", "tags.customer", "test"),
				),
			},
		},
	})
}

func testAccCheckDatastoreDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "opennebula_datastore":

			datastoreID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			dc := controller.Datastore(int(datastoreID))
			datastore, _ := dc.Info(false)
			if datastore != nil {
				return fmt.Errorf("Expected datastore %s to have been destroyed", rs.Primary.ID)
			}
		default:
		}
	}

	return nil
}

var testAccDatastoreConfigBasic = `
resource "opennebula_datastore" "example" {
	name = "test"
	type = "image"

	custom {
		datastore = "dummy"
		transfer = "dummy"
	}

	tags = {
		test = "test"
		environment = "example"
	  }
  }
`

var testAccDatastoreConfigUpdate = `
resource "opennebula_datastore" "example" {
	name = "test-updated"
	type = "image"

	custom {
		datastore = "dummy"
		transfer = "dummy"
	}

	tags = {
		environment = "example-updated"
		customer = "test"
	  }
  }
`
