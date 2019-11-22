package opennebula

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"strconv"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

func TestAccSecurityGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccSecurityGroupConfigBasic,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "name", "testsg"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "permissions", "642"),
					testAccSecurityGroupRule(0, "PROTOCOL", "ALL"),
					testAccSecurityGroupRule(0, "RULE_TYPE", "OUTBOUND"),
					testAccSecurityGroupRule(1, "PROTOCOL", "TCP"),
					testAccSecurityGroupRule(1, "RULE_TYPE", "INBOUND"),
					testAccSecurityGroupRule(1, "RANGE", "22"),
					testAccSecurityGroupRule(2, "PROTOCOL", "ICMP"),
					testAccSecurityGroupRule(2, "RULE_TYPE", "INBOUND"),
				),
			},
			{
				Config:             testAccSecurityGroupConfigUpdate,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "name", "renamedsg"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "permissions", "660"),
					testAccSecurityGroupRule(0, "PROTOCOL", "ALL"),
					testAccSecurityGroupRule(0, "RULE_TYPE", "OUTBOUND"),
					testAccSecurityGroupRule(1, "PROTOCOL", "TCP"),
					testAccSecurityGroupRule(1, "RULE_TYPE", "INBOUND"),
					testAccSecurityGroupRule(1, "RANGE", "80"),
					testAccSecurityGroupRule(2, "PROTOCOL", "ICMP"),
					testAccSecurityGroupRule(2, "RULE_TYPE", "INBOUND"),
				),
			},
		},
	})
}

func testAccCheckSecurityGroupDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		sgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		sgc := controller.SecurityGroup(int(sgID))
		// Get Security Group Info
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		sg, _ := sgc.Info(false)
		if sg != nil {
			return fmt.Errorf("Expected security group %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

func testAccSecurityGroupRule(ruleidx int, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			sgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			sgc := controller.SecurityGroup(int(sgID))
			// Get Security Group Info
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			sg, _ := sgc.Info(false)
			if sg == nil {
				return fmt.Errorf("Expected Security Group %s to exist when checking permissions", rs.Primary.ID)
			}
			sgrules := generateSecurityGroupMapFromStructs(sg.Template.GetRules())

			var found bool

			for i, rule := range sgrules {
				if i == ruleidx {
					if rule[key] != nil && rule[key].(string) != value {
						return fmt.Errorf("Expected %s = %s for rule ID %d, got %s = %s", key, value, ruleidx, key, rule[key].(string))
					}
					found = true
				}
			}

			if !found {
				return fmt.Errorf("rule id %d with %s = %s does not exist", ruleidx, key, value)
			}

		}

		return nil
	}
}

var testAccSecurityGroupConfigBasic = `
resource "opennebula_security_group" "mysecgroup" {
    name = "testsg"
    description = "Terraform security group"
    permissions = "642"
    rule {
        protocol = "ALL"
        rule_type = "OUTBOUND"
    }
    rule {
        protocol = "TCP"
        rule_type = "INBOUND"
        range = "22"
    }
    rule {
        protocol = "ICMP"
        rule_type = "INBOUND"
    }
}
`

var testAccSecurityGroupConfigUpdate = `
resource "opennebula_security_group" "mysecgroup" {
    name = "renamedsg"
    description = "Terraform security group"
    permissions = "660"
    rule {
        protocol = "ALL"
        rule_type = "OUTBOUND"
    }
    rule {
        protocol = "TCP"
        rule_type = "INBOUND"
        range = "80"
    }
    rule {
        protocol = "ICMP"
        rule_type = "INBOUND"
    }
}
`
