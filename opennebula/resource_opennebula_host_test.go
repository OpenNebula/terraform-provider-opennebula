package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHostConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_host.test1", "name", "test-custom"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "type", "custom"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_host.test1", "custom.*", map[string]string{
						"virtualization": "dummy",
						"information":    "dummy",
					}),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.environment", "example"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.test", "test"),
				),
			},
			{
				Config: testAccHostConfigAddOvercommit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_host.test1", "name", "test-updated"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "type", "custom"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_host.test1", "custom.*", map[string]string{
						"virtualization": "dummy",
						"information":    "dummy",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_host.test1", "overcommit.*", map[string]string{
						"cpu":    "3200",
						"memory": "1048576",
					}),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.environment", "example-updated"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.customer", "test"),
				),
			},
			{
				Config: testAccHostConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_host.test1", "name", "test-updated"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "type", "custom"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_host.test1", "custom.*", map[string]string{
						"virtualization": "dummy",
						"information":    "dummy",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_host.test1", "overcommit.*", map[string]string{
						"cpu":    "3300",
						"memory": "1148576",
					}),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.environment", "example-updated"),
					resource.TestCheckResourceAttr("opennebula_host.test1", "tags.customer", "test"),
				),
			},
		},
	})
}

func testAccCheckHostDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		hostID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
		ic := controller.Host(int(hostID))
		// Get Host Info
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		host, _ := ic.Info(false)
		if host != nil {
			return fmt.Errorf("Expected host %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

var testAccHostConfig = `
resource "opennebula_host" "test1" {
	name       = "test-custom"
	type       = "custom"

	custom {
		virtualization = "dummy"
		information = "dummy"
	}
  
	tags = {
	  test = "test"
	  environment = "example"
	}
  }
`

var testAccHostConfigAddOvercommit = `
resource "opennebula_host" "test1" {
	name       = "test-updated"
	type       = "custom"

	custom {
		virtualization = "dummy"
		information = "dummy"
	}
  
	overcommit {
	  cpu = 3200        # 32 cores
	  memory = 1048576  # 1 Gb
	}
  
	tags = {
	  environment = "example-updated"
	  customer = "test"
	}
  }
`

var testAccHostConfigUpdate = `
resource "opennebula_host" "test1" {
	name       = "test-updated"
	type       = "custom"

	custom {
		virtualization = "dummy"
		information = "dummy"
	}

	overcommit {
	  cpu = 3300
	  memory = 1148576
	}

	tags = {
	  environment = "example-updated"
	  customer = "test"
	}
  }
`
