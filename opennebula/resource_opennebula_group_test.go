package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

func TestAccGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccGroupConfigBasic,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttr("opennebula_group.group", "delete_on_destruction", "false"),
				),
			},
			{
				Config:             testAccGroupConfigUpdate,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttr("opennebula_group.group", "delete_on_destruction", "true"),
				),
			},
			{
				Config:             testAccGroupLigh,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group2", "name", "noquotas"),
					resource.TestCheckResourceAttr("opennebula_group.group2", "delete_on_destruction", "true"),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		groupID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		gc := controller.Group(int(groupID))
		// Get Group Info
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		group, _ := gc.Info(false)
		if group != nil {
			return fmt.Errorf("Expected group %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

var testAccGroupConfigBasic = `
resource "opennebula_group" "group" {
  name = "iamgroup"
  template = <<EOF
    SUNSTONE = [
      DEFAULT_VIEW = "cloud",
      GROUP_ADMIN_DEFAULT_VIEW = "groupadmin",
      GROUP_ADMIN_VIEWS = "groupadmin",
      VIEWS = "cloud"
    ]
    EOF
    delete_on_destruction = false
    quotas {
        datastore_quotas {
            id = 1
            images = 3
            size = 100
        }
        vm_quotas {
            cpu = 4
            memory = 8192
        }
    }
}
`

var testAccGroupConfigUpdate = `
resource "opennebula_group" "group" {
  name = "iamgroup"
  template = <<EOF
    SUNSTONE = [
      DEFAULT_VIEW = "cloud",
      GROUP_ADMIN_DEFAULT_VIEW = "groupadmin",
      GROUP_ADMIN_VIEWS = "cloud",
      VIEWS = "cloud"
    ]
    EOF
    delete_on_destruction = true
    quotas {
        datastore_quotas {
            id = 1
            images = 4
            size = 100
        }
        vm_quotas {
            cpu = 4
            memory = 8192
        }
    }
}
`

var testAccGroupLigh = `
resource "opennebula_group" "group" {
  name = "iamgroup"
  template = <<EOF
    SUNSTONE = [
      DEFAULT_VIEW = "cloud",
      GROUP_ADMIN_DEFAULT_VIEW = "groupadmin",
      GROUP_ADMIN_VIEWS = "cloud",
      VIEWS = "cloud"
    ]
    EOF
    delete_on_destruction = true
    quotas {
        datastore_quotas {
            id = 1
            images = 4
            size = 100
        }
        vm_quotas {
            cpu = 4
            memory = 8192
        }
    }
}

resource "opennebula_group" "group2" {
  name = "noquotas"
  template = <<EOF
    SUNSTONE = [
      DEFAULT_VIEW = "cloud",
      GROUP_ADMIN_DEFAULT_VIEW = "groupadmin",
      GROUP_ADMIN_VIEWS = "cloud",
      VIEWS = "cloud"
    ]
    EOF
    delete_on_destruction = true
}
`
