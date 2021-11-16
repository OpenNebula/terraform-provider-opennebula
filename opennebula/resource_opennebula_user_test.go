package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.datastore_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.datastore_quotas.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.datastore_quotas.0.images", "3"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.datastore_quotas.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.image_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.network_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.vm_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.vm_quotas.2832483756.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.4169128061.vm_quotas.2832483756.memory", "8192"),
				),
			},
			{
				Config: testAccUserConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_user.user", "name", "iamuser"),
					resource.TestCheckResourceAttr("opennebula_user.user", "password", "p@ssw0rd2"),
					resource.TestCheckResourceAttr("opennebula_user.user", "auth_driver", "core"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.datastore_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.datastore_quotas.0.id", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.datastore_quotas.0.images", "4"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.datastore_quotas.0.size", "100"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.image_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.network_quotas.#", "0"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.vm_quotas.#", "1"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.vm_quotas.2832483756.cpu", "4"),
					resource.TestCheckResourceAttr("opennebula_user.user", "quotas.261273647.vm_quotas.2832483756.memory", "8192"),
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
