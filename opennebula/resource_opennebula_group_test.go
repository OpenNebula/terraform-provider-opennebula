package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

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
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.datastore_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.datastore_quotas.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.datastore_quotas.0.images", "3"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.datastore_quotas.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.image_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.network_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.vm_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.vm_quotas.2832483756.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.4169128061.vm_quotas.2832483756.memory", "8192"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.4128055932.default_view", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.4128055932.group_admin_default_view", "groupadmin"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.4128055932.group_admin_views", "groupadmin"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.4128055932.views", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey1", "testvalue1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey2", "testvalue2"),
				),
			},
			{
				Config: testAccGroupConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttr("opennebula_group.group", "delete_on_destruction", "true"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.datastore_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.datastore_quotas.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.datastore_quotas.0.images", "4"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.datastore_quotas.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.image_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.network_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.vm_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.vm_quotas.2832483756.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.261273647.vm_quotas.2832483756.memory", "8192"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.1904779058.default_view", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.1904779058.group_admin_default_view", "groupadmin"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.1904779058.group_admin_views", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.1904779058.views", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey2", "testvalue2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey3", "testvalue3"),
				),
			},
			{
				Config: testAccGroupWithUser,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_user.user", "name", "iamuser"),
					resource.TestCheckResourceAttrSet("opennebula_user.user", "primary_group"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.1904779058.default_view", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.1904779058.group_admin_default_view", "groupadmin"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.1904779058.group_admin_views", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.1904779058.views", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey2", "testvalue2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey3", "testvalue3"),
				),
			},
			{
				Config: testAccGroupWithGroupAdmin,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("opennebula_group_admins.admins", "group_id"),
					resource.TestCheckResourceAttr("opennebula_group_admins.admins", "users_ids.#", "1"),
				),
			},
			{
				Config: testAccGroupLigh,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group2", "name", "noquotas"),
					resource.TestCheckResourceAttr("opennebula_group.group2", "delete_on_destruction", "true"),
					resource.TestCheckResourceAttr("opennebula_group.group2", "sunstone.1904779058.default_view", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group2", "sunstone.1904779058.group_admin_default_view", "groupadmin"),
					resource.TestCheckResourceAttr("opennebula_group.group2", "sunstone.1904779058.group_admin_views", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group2", "sunstone.1904779058.views", "cloud"),
					resource.TestCheckResourceAttr("opennebula_group.group2", "tags.%", "0"),
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

var testAccGroupUser = `
resource "opennebula_user" "user" {
	name          = "iamuser"
	password      = "password"
	auth_driver   = "core"
	primary_group = opennebula_group.group.id
  }
`

var testAccGroupConfigBasic = `
resource "opennebula_group" "group" {
    name = "iamgroup"
    sunstone {
      default_view = "cloud"
      group_admin_default_view = "groupadmin"
      group_admin_views = "groupadmin"
      views = "cloud"
	}
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
	tags = {
		testkey1 = "testvalue1"
		testkey2 = "testvalue2"
	}

	lifecycle {
		ignore_changes = [
			template,
		]
	}
}
`

var testAccGroupConfigUpdate = `
resource "opennebula_group" "group" {
    name = "iamgroup"
	sunstone {
		default_view = "cloud"
		group_admin_default_view = "groupadmin"
		group_admin_views = "cloud"
		views = "cloud"
	}
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
	tags = {
		testkey2 = "testvalue2"
		testkey3 = "testvalue3"
	}

	lifecycle {
		ignore_changes = [
			template,
		]
	}
}
`

var testAccGroupWithUser = testAccGroupConfigUpdate + testAccGroupUser

var testAccGroupWithGroupAdmin = testAccGroupWithUser + `
resource "opennebula_group_admins" "admins" {
	group_id = opennebula_group.group.id
	users_ids = [
	  opennebula_user.user.id
	]
  }
`

var testAccGroupLigh = `
resource "opennebula_group" "group" {
    name = "iamgroup"
	sunstone {
		default_view = "cloud"
		group_admin_default_view = "groupadmin"
		group_admin_views = "cloud"
		views = "cloud"
	}
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

	lifecycle {
		ignore_changes = [
			template,
		]
	}
}

resource "opennebula_group" "group2" {
    name = "noquotas"
	sunstone {
		default_view = "cloud"
		group_admin_default_view = "groupadmin"
		group_admin_views = "cloud"
		views = "cloud"
	}
    delete_on_destruction = true

	lifecycle {
		ignore_changes = [
			template,
		]
	}
}
`
