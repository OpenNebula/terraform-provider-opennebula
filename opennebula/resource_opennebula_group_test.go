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
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.0.%", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "sunstone.*", map[string]string{
						"default_view":             "cloud",
						"group_admin_default_view": "groupadmin",
						"group_admin_views":        "groupadmin",
						"views":                    "cloud",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "opennebula.*", map[string]string{
						"default_image_persistent": "YES",
						"api_list_order":           "ASC",
					}),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey1", "testvalue1"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey2", "testvalue2"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey1": "testvalue1",
						"elements.testkey2": "testvalue2",
					}),
					resource.TestCheckResourceAttr("opennebula_group_quotas.datastore", "datastore.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.datastore", "datastore.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.datastore", "datastore.0.images", "3"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.datastore", "datastore.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.vm", "vm.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.vm", "vm.0.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.vm", "vm.0.memory", "8192"),
				),
			},
			{
				Config: testAccGroupConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_group.group", "name", "iamgroup"),
					resource.TestCheckResourceAttr("opennebula_group.group", "sunstone.0.%", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "sunstone.*", map[string]string{
						"default_view":             "cloud",
						"group_admin_default_view": "groupadmin",
						"group_admin_views":        "cloud",
						"views":                    "cloud",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "opennebula.*", map[string]string{
						"api_list_order": "DESC",
					}),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey2", "testvalue2"),
					resource.TestCheckResourceAttr("opennebula_group.group", "tags.testkey3", "testvalue3"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey2": "testvalue2",
						"elements.testkey3": "testvalue3",
					}),
					resource.TestCheckResourceAttr("opennebula_group_quotas.datastore", "datastore.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.datastore", "datastore.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.datastore", "datastore.0.images", "4"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.datastore", "datastore.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.vm", "vm.#", "1"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.vm", "vm.0.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_group_quotas.vm", "vm.0.memory", "8192"),
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
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_group.group", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey2": "testvalue2",
						"elements.testkey3": "testvalue3",
					}),
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
					resource.TestCheckResourceAttr("opennebula_group.group2", "template_section.#", "0"),
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
	opennebula {
	  default_image_persistent = "YES"
	  api_list_order = "ASC"
	}

	tags = {
		testkey1 = "testvalue1"
		testkey2 = "testvalue2"
	}

	template_section {
		name = "test_vec_key"
		elements = {
			testkey1 = "testvalue1"
			testkey2 = "testvalue2"
		}
	}

	lifecycle {
		ignore_changes = [
		  "quotas"
		]
	  }
}

resource "opennebula_group_quotas" "datastore" {
	group_id = opennebula_group.group.id
	datastore {
		id = 1
		images = 3
		size = 100
	}
}

resource "opennebula_group_quotas" "vm" {
	group_id = opennebula_group.group.id
	vm {
		cpu = 4
		memory = 8192
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
	opennebula {
		api_list_order = "DESC"
	}

	tags = {
		testkey2 = "testvalue2"
		testkey3 = "testvalue3"
	}

	template_section {
		name = "test_vec_key"
		elements = {
			testkey2 = "testvalue2"
			testkey3 = "testvalue3"
		}
	}

	lifecycle {
		ignore_changes = [
		  "quotas"
		]
	  }
}

resource "opennebula_group_quotas" "datastore" {
	group_id = opennebula_group.group.id
	datastore {
		id = 1
		images = 4
		size = 100
	}
}

resource "opennebula_group_quotas" "vm" {
	group_id = opennebula_group.group.id
	vm {
		cpu = 4
		memory = 8192
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

	tags = {
		testkey2 = "testvalue2"
		testkey3 = "testvalue3"
	}

	template_section {
		name = "test_vec_key"
		elements = {
			testkey2 = "testvalue2"
			testkey3 = "testvalue3"
		}
	}

	lifecycle {
		ignore_changes = [
		  "quotas"
		]
	  }
}

resource "opennebula_group_quotas" "datastore" {
	group_id = opennebula_group.group.id
	datastore {
		id = 1
		images = 4
		size = 100
	}
}

resource "opennebula_group_quotas" "vm" {
	group_id = opennebula_group.group.id
	vm {
		cpu = 4
		memory = 8192
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

	lifecycle {
		ignore_changes = [
		  "quotas"
		]
	  }
}

resource "opennebula_group_quotas" "datastore" {
	group_id = opennebula_group.group.id
	datastore {
		id = 1
		images = 4
		size = 100
	}
}

resource "opennebula_group_quotas" "vm" {
	group_id = opennebula_group.group.id
	vm {
		cpu = 4
		memory = 8192
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
