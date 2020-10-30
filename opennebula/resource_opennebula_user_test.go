package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
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
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.datastore_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.datastore_quotas.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.datastore_quotas.0.images", "3"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.datastore_quotas.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.image_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.network_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.vm_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.vm_quotas.947884945.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.vm_quotas.947884945.memory", "8192"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.vm_quotas.947884945.running_cpu", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.vm_quotas.947884945.running_memory", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.vm_quotas.947884945.running_vms", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.vm_quotas.947884945.system_disk_size", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.2413004587.vm_quotas.947884945.vms", "0"),
				),
			},
			{
				Config: testAccUserConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_user.user", "name", "iamuser"),
					resource.TestCheckResourceAttr("opennebula_user.user", "password", "p@ssw0rd2"),
					resource.TestCheckResourceAttr("opennebula_user.user", "auth_driver", "core"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.datastore_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.datastore_quotas.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.datastore_quotas.0.images", "4"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.datastore_quotas.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.image_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.network_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.vm_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.vm_quotas.947884945.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.vm_quotas.947884945.memory", "8192"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.vm_quotas.947884945.running_cpu", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.vm_quotas.947884945.running_memory", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.vm_quotas.947884945.running_vms", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.vm_quotas.947884945.system_disk_size", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.267449028.vm_quotas.947884945.vms", "0"),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		userID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
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

var testAccUserConfigUpdate = `
resource "opennebula_user" "user" {
  name = "iamuser"
  password = "p@ssw0rd2"
  auth_driver = "core"
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
