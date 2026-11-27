package opennebula

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccACL(t *testing.T) {
	invalidUserErr, _ := regexp.Compile("ID String something malformed")
	invalidResourceErr, _ := regexp.Compile("resource 'a' malformed")
	invalidRightsErr, _ := regexp.Compile("right 'aa' does not exist")
	invalidZoneErr, _ := regexp.Compile("ID String bad malformed")
	var deletedID int
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testACLConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_acl.acl_foo", "user", "@1"),
					resource.TestCheckResourceAttrWith("opennebula_acl.acl_foo", "id", func(value string) error {
						var err error
						deletedID, err = strconv.Atoi(value)
						return err
					}),
				),
			},
			{
				PreConfig: func() {
					controller := testAccProvider.Meta().(*Configuration).Controller
					if err := controller.ACLs().DeleteRule(deletedID); err != nil {
						t.Fatalf("delete ACL outside Terraform: %v", err)
					}
				},
				Config: testACLConfigBasic,
				Check: resource.TestCheckResourceAttrWith("opennebula_acl.acl_foo", "id", func(value string) error {
					if value == strconv.Itoa(deletedID) {
						return fmt.Errorf("ACL ID %s did not change after external deletion", value)
					}
					return nil
				}),
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
			{
				Config:      testACLInvalidZone,
				ExpectError: invalidZoneErr,
			},
		},
	})
}

func testAccCheckACLDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller
	acls, err := controller.ACLs().Info()

	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		id, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
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
  zone = "#0"
}
`

var testACLConfigReplace = `
resource "opennebula_acl" "acl_foo" {
  user = "@0"
  resource = "HOST+CLUSTER+DATASTORE/*"
  rights = "USE+MANAGE+ADMIN"
  zone = "*"
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

var testACLInvalidZone = `
resource "opennebula_acl" "acl_invalid" {
  user = "@1"
  resource = "HOST+CLUSTER+DATASTORE/*"
  rights = "USE+MANAGE+ADMIN"
  zone = "bad"
}
`
