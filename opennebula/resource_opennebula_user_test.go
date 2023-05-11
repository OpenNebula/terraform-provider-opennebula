package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_user.user", "name", "iamuser"),
					resource.TestCheckResourceAttr("opennebula_user.user", "password", "p@ssw0rd"),
					resource.TestCheckResourceAttr("opennebula_user.user", "auth_driver", "core"),
					resource.TestCheckResourceAttr("opennebula_user.user", "ssh_public_key", "xxx"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.datastore", "datastore.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.datastore", "datastore.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.datastore", "datastore.0.images", "3"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.datastore", "datastore.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.vm", "vm.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.vm", "vm.0.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.vm", "vm.0.memory", "8192"),
					resource.TestCheckResourceAttr("opennebula_user.user", "tags.testkey1", "testvalue1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "tags.testkey2", "testvalue2"),
				),
			},
			{
				Config: testAccUserConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_user.user", "name", "iamuser"),
					resource.TestCheckResourceAttr("opennebula_user.user", "password", "p@ssw0rd2"),
					resource.TestCheckResourceAttr("opennebula_user.user", "auth_driver", "core"),
					resource.TestCheckResourceAttr("opennebula_user.user", "ssh_public_key", "xxx"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.datastore", "datastore.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.datastore", "datastore.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.datastore", "datastore.0.images", "4"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.datastore", "datastore.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.vm", "vm.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.vm", "vm.0.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_user_quotas.vm", "vm.0.memory", "8192"),
					resource.TestCheckResourceAttr("opennebula_user.user", "tags.testkey2", "testvalue2"),
					resource.TestCheckResourceAttr("opennebula_user.user", "tags.testkey3", "testvalue3"),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		userID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
		uc := controller.User(int(userID))
		// Get User Info
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		user, _ := uc.Info(false)
		if user != nil {
			return fmt.Errorf("Expected user %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

var testAccUserConfigBasic = `
resource "opennebula_user" "user" {
  name = "iamuser"
  password = "p@ssw0rd"
  auth_driver = "core"
  ssh_public_key = "xxx"

  tags = {
	testkey1 = "testvalue1"
	testkey2 = "testvalue2"
  }

  lifecycle {
    ignore_changes = [
      "quotas"
    ]
  }
}

resource "opennebula_user_quotas" "datastore" {
  user_id = opennebula_user.user.id
  datastore {
    id = 1
    images = 3
    size = 100
  }
}

resource "opennebula_user_quotas" "vm" {
	user_id = opennebula_user.user.id
	vm {
	  cpu = 4
	  memory = 8192
	}
  }
`

var testAccUserConfigUpdate = `
resource "opennebula_user" "user" {
  name = "iamuser"
  password = "p@ssw0rd2"
  auth_driver = "core"
  ssh_public_key = "xxx"

  tags = {
	testkey2 = "testvalue2"
	testkey3 = "testvalue3"
  }

  lifecycle {
    ignore_changes = [
      "quotas"
    ]
  }
}

resource "opennebula_user_quotas" "datastore" {
  user_id = opennebula_user.user.id
  datastore {
    id = 1
    images = 4
    size = 100
  }
}

resource "opennebula_user_quotas" "vm" {
	user_id = opennebula_user.user.id
	vm {
	  cpu = 4
	  memory = 8192
	}
  }
`
