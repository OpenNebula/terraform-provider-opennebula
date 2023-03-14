package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMarketplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMarketplaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMarketplaceConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "name", "testmp"),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "permissions", "642"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_marketplace.example", "s3.*", map[string]string{
						"type":              "aws",
						"access_key_id":     "testkey",
						"secret_access_key": "testsecretkey",
						"region":            "somewhere",
						"bucket":            "bucket1",
					}),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "tags.customer", "test"),
				),
			},
			{
				Config: testAccMarketplaceConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "name", "renamedmp"),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "permissions", "642"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_marketplace.example", "s3.*", map[string]string{
						"type":              "aws",
						"access_key_id":     "testkey",
						"secret_access_key": "testsecretkey",
						"region":            "somewhere",
						"bucket":            "bucket2",
					}),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "tags.%", "3"),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_marketplace.example", "tags.version", "2"),
				),
			},
		},
	})
}

func testAccCheckMarketplaceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		mpID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
		mpc := controller.MarketPlace(int(mpID))
		mp, _ := mpc.Info(false)
		if mp != nil {
			return fmt.Errorf("Expected marketplace %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

var testAccMarketplaceConfigBasic = `
resource "opennebula_marketplace" "example" {
    name = "testmp"
    description = "Terraform marketplace"
    permissions = "642"

	s3 {
		type = "aws"
		access_key_id = "testkey"
		secret_access_key = "testsecretkey"
		region = "somewhere"
		bucket = "bucket1"
	}

    tags = {
      env = "prod"
      customer = "test"
    }
}
`

var testAccMarketplaceConfigUpdate = `
resource "opennebula_marketplace" "example" {
    name = "renamedmp"
    description = "Terraform marketplace"
    permissions = "642"

	s3 {
		type = "aws"
		access_key_id = "testkey"
		secret_access_key = "testsecretkey"
		region = "somewhere"
		bucket = "bucket2"
	}

    tags = {
      env = "dev"
      customer = "test"
      version = "2"
    }
}
`
