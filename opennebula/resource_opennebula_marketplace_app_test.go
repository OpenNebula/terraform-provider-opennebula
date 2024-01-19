package opennebula

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMarketplaceAppliance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		ErrorCheck: func(err error) error {
			// ignore acceptance test error for ONE 5.12
			match, _ := regexp.Match(`Other type than IMAGE not supported before OpenNebula 6.0`, []byte(err.Error()))
			if match {
				return nil
			}
			return err
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMarketplaceApplianceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMarketplaceApplianceConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "name", "test-app"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "type", "VMTEMPLATE"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "description", "this is an appilance"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "tags.custom1", "value1"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "permissions", "642"),
				),
			},
			{
				Config: testAccMarketplaceApplianceConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "name", "test-renamed-app"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "type", "VMTEMPLATE"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "description", "this is an appliance"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "tags.custom1", "value2"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "tags.custom3", "value3"),
					resource.TestCheckResourceAttr("opennebula_marketplace_appliance.example", "permissions", "642"),
				),
			},
		},
	})
}

func testAccCheckMarketplaceApplianceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "opennebula_marketplace":
			mpID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			mp := controller.MarketPlace(int(mpID))
			_, err := mp.Info(false)
			if !NoExists(err) {
				return fmt.Errorf("Expected marketplace %s to have been destroyed", rs.Primary.ID)

			}
		case "opennebula_marketplace_appliance":
			mpID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			mpc := controller.MarketPlaceApp(int(mpID))
			_, err := mpc.Info(false)
			if !NoExists(err) {
				return fmt.Errorf("Expected appliance %s to have been destroyed", rs.Primary.ID)

			}
		}
	}

	return nil
}

var testAccMarketplaceApplianceConfigBasic = `

resource "opennebula_marketplace" "example" {
    name = "test-market"
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

resource "opennebula_marketplace_appliance" "example" {
	name = "test-app"
	market_id = opennebula_marketplace.example.id
    permissions = "642"
	type = "VMTEMPLATE"
	description = "this is an appilance"
	version = "0.2.0"
  
	tags = {
	  custom1 = "value1"
	}
  }
`

var testAccMarketplaceApplianceConfigUpdate = `
resource "opennebula_marketplace" "example" {
    name = "test-market"
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


resource "opennebula_marketplace_appliance" "example" {
	name = "test-renamed-app"
	market_id = opennebula_marketplace.example.id
    permissions = "642"
	type = "VMTEMPLATE"
	description = "this is an appliance"
	version = "1.0.0"
  
	tags = {
	  custom1 = "value2"
	  custom3 = "value3"
	}
  }
`
