package opennebula

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccACL(t *testing.T) {
	invalidUserErr, _ := regexp.Compile("ID String something malformed")
	invalidResourceErr, _ := regexp.Compile("Resource 'a' malformed")
	invalidRightsErr, _ := regexp.Compile("Right 'aa' does not exist.")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testACLConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_acl.acl_foo", "user", "@1"),
				),
			},
			{
				Config: testACLConfigReplace,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_acl.acl_foo", "user", "@0"),
				),
			},
			{
				Config:      testACLInvalidUser,
				ExpectError: invalidUserErr,
			},
			{
				Config:      testACLInvalidResource,
				ExpectError: invalidResourceErr,
			},
			{
				Config:      testACLInvalidRights,
				ExpectError: invalidRightsErr,
			},
		},
	})
}

func testAccCheckACLDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)
	acls, err := controller.ACLs().Info()

	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		id, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		for _, acl := range acls.ACLs {
			if int(id) == acl.ID {
				return fmt.Errorf("Expected group %s to have been destroyed", rs.Primary.ID)
			}
		}
	}

	return nil
}

var testACLConfigBasic = `
resource "opennebula_acl" "acl_foo" {
  user = "@1"
  resource = "HOST+CLUSTER+DATASTORE/*"
  rights = "USE+MANAGE+ADMIN"
}
`

var testACLConfigReplace = `
resource "opennebula_acl" "acl_foo" {
  user = "@0"
  resource = "HOST+CLUSTER+DATASTORE/*"
  rights = "USE+MANAGE+ADMIN"
}
`

var testACLInvalidUser = `
resource "opennebula_acl" "acl_invalid" {
  user = "something"
  resource = "HOST+CLUSTER+DATASTORE/*"
  rights = "USE+MANAGE+ADMIN"
}
`

var testACLInvalidResource = `
resource "opennebula_acl" "acl_invalid" {
  user = "@1"
  resource = "a"
  rights = "USE+MANAGE+ADMIN"
}
`

var testACLInvalidRights = `
resource "opennebula_acl" "acl_invalid" {
  user = "@1"
  resource = "HOST+CLUSTER+DATASTORE/*"
  rights = "aa"
}
`
