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
				Config: testAccGroupConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttr("opennebula_group.group", "delete_on_destruction", "false"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.datastore_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.datastore_quotas.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.datastore_quotas.0.images", "3"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.datastore_quotas.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.image_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.network_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.vm_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.vm_quotas.947884945.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.vm_quotas.947884945.memory", "8192"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.vm_quotas.947884945.running_cpu", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.vm_quotas.947884945.running_memory", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.vm_quotas.947884945.running_vms", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.vm_quotas.947884945.system_disk_size", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.2413004587.vm_quotas.947884945.vms", "0"),
				),
			},
			{
				Config: testAccGroupConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttr("opennebula_group.group", "delete_on_destruction", "true"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.datastore_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.datastore_quotas.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.datastore_quotas.0.images", "4"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.datastore_quotas.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.image_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.network_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.vm_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.vm_quotas.947884945.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.vm_quotas.947884945.memory", "8192"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.vm_quotas.947884945.running_cpu", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.vm_quotas.947884945.running_memory", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.vm_quotas.947884945.running_vms", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.vm_quotas.947884945.system_disk_size", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.267449028.vm_quotas.947884945.vms", "0"),
				),
			},
			{
				Config: testAccGroupLigh,
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
