package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccSecurityGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "name", "testsg"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.#", "2"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.0.protocol", "ALL"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.0.rule_type", "OUTBOUND"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.1.protocol", "ICMP"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.1.rule_type", "INBOUND"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "tags.customer", "test"),
				),
			},
			{
				Config: testAccSecurityGroupConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "name", "renamedsg"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.#", "3"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.0.protocol", "ALL"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.0.rule_type", "OUTBOUND"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.1.protocol", "TCP"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.1.rule_type", "INBOUND"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.1.range", "80"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.2.protocol", "ICMP"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "rule.2.rule_type", "INBOUND"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "tags.%", "3"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_security_group.mysecgroup", "tags.version", "2"),
				),
			},
		},
	})
}

func testAccCheckSecurityGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		sgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
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
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			sgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
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
        protocol = "ICMP"
        rule_type = "INBOUND"
    }
    tags = {
      env = "prod"
      customer = "test"
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
    tags = {
      env = "dev"
      customer = "test"
      version = "2"
    }
}
`
