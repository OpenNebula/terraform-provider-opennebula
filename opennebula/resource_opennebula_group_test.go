package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "quotas.*", map[string]string{
						"datastore_quotas.#":        "1",
						"datastore_quotas.0.id":     "1",
						"datastore_quotas.0.images": "3",
						"datastore_quotas.0.size":   "100",
						"image_quotas.#":            "0",
						"network_quotas.#":          "0",
						"vm_quotas.#":               "1",
						"vm_quotas.0.cpu":           "4",
						"vm_quotas.0.memory":        "8192",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "sunstone.*", map[string]string{
						"default_view":             "cloud",
						"group_admin_default_view": "groupadmin",
						"group_admin_views":        "groupadmin",
						"views":                    "cloud",
					}),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey1", "testvalue1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey2", "testvalue2"),
				),
			},
			{
				Config: testAccGroupConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttr("opennebula_group.group", "quotas.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "quotas.*", map[string]string{
						"datastore_quotas.#":        "1",
						"datastore_quotas.0.id":     "1",
						"datastore_quotas.0.images": "4",
						"datastore_quotas.0.size":   "100",
						"image_quotas.#":            "0",
						"network_quotas.#":          "0",
						"vm_quotas.#":               "1",
						"vm_quotas.0.cpu":           "4",
						"vm_quotas.0.memory":        "8192",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "sunstone.*", map[string]string{
						"default_view":             "cloud",
						"group_admin_default_view": "groupadmin",
						"group_admin_views":        "cloud",
						"views":                    "cloud",
					}),
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
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "sunstone.*", map[string]string{
						"default_view":             "cloud",
						"group_admin_default_view": "groupadmin",
						"group_admin_views":        "cloud",
						"views":                    "cloud",
					}),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey2", "testvalue2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey3", "testvalue3"),
				),
			},
			{
				Config: testAccGroupWithGroupAdmin,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_user.user", "name", "iamuser"),
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttrSet("opennebula_group_admins.admins", "group_id"),
					resource.TestCheckResourceAttr("opennebula_group_admins.admins", "users_ids.#", "1"),
				),
			},
			{
				Config: testAccGroupWithGroupAdminUserRenamed,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_user.user", "name", "iamuser_renamed"),
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttrSet("opennebula_group_admins.admins", "group_id"),
					resource.TestCheckResourceAttr("opennebula_group_admins.admins", "users_ids.#", "1"),
				),
			},
			{
				Config: testAccGroupWithGroupAdminGroupRenamed,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_user.user", "name", "iamuser_renamed"),
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup_renamed"),
					resource.TestCheckResourceAttrSet("opennebula_group_admins.admins", "group_id"),
					resource.TestCheckResourceAttr("opennebula_group_admins.admins", "users_ids.#", "1"),
				),
			},
			{
				Config: testAccGroupLigh,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group2", "name", "noquotas"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "sunstone.*", map[string]string{
						"default_view":             "cloud",
						"group_admin_default_view": "groupadmin",
						"group_admin_views":        "cloud",
						"views":                    "cloud",
					}),
					resource.TestCheckResourceAttr("opennebula_group.group2", "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		groupID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
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

var testAccGroupUserRenamed = `
resource "opennebula_user" "user" {
	name          = "iamuser_renamed"
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
}
`

var testAccGroupConfigGroupRenamed = `
resource "opennebula_group" "group" {
    name = "iamgroup_renamed"
	sunstone {
		default_view = "cloud"
		group_admin_default_view = "groupadmin"
		group_admin_views = "cloud"
		views = "cloud"
	}
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

var testAccGroupWithGroupAdminUserRenamed = testAccGroupConfigUpdate + testAccGroupUserRenamed + `
resource "opennebula_group_admins" "admins" {
	group_id = opennebula_group.group.id
	users_ids = [
	  opennebula_user.user.id
	]
  }
`

var testAccGroupWithGroupAdminGroupRenamed = testAccGroupConfigGroupRenamed + testAccGroupUserRenamed + `
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
	sunstone {
		default_view = "cloud"
		group_admin_default_view = "groupadmin"
		group_admin_views = "cloud"
		views = "cloud"
	}
}
`
