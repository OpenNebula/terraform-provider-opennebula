package opennebula

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_cluster.example", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "hosts.#", "1"),
					resource.TestCheckTypeSetElemAttr("opennebula_cluster.example", "hosts.*", "0"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "datastores.#", "1"),
					resource.TestCheckTypeSetElemAttr("opennebula_cluster.example", "datastores.*", "2"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "virtual_networks.#", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "tags.environment", "example"),
				),
			},
			{
				Config: testAccClusterConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_cluster.example", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "hosts.#", "1"),
					resource.TestCheckTypeSetElemAttr("opennebula_cluster.example", "hosts.*", "0"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "datastores.#", "1"),
					resource.TestCheckTypeSetElemAttr("opennebula_cluster.example", "datastores.*", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "virtual_networks.#", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "tags.environment", "updated"),
				),
			},
			{
				Config: testAccClusterConfigEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_cluster.example", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "hosts.#", "0"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "datastores.#", "0"),
					resource.TestCheckResourceAttr("opennebula_cluster.example", "virtual_networks.#", "0"),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "opennebula_cluster":

			clusterID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			gc := controller.Cluster(int(clusterID))
			cluster, _ := gc.Info()
			if cluster != nil {
				return fmt.Errorf("Expected cluster %s to have been destroyed", rs.Primary.ID)
			}
		default:
		}
	}

	return nil
}

var testAccClusterVnet = `
resource "opennebula_virtual_network" "test" {
	name            = "test"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500

	security_groups = [0]
	clusters = [0]
	permissions = "660"
	group = "users"

	lifecycle {
	  ignore_changes = [ar, hold_ips]
	}
}

resource "opennebula_virtual_network" "test2" {
	name            = "test2"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500

	security_groups = [0]
	clusters = [0]
	permissions = "660"
	group = "users"

	lifecycle {
	  ignore_changes = [ar, hold_ips]
	}
}`

var testAccClusterConfigEmpty = testAccClusterVnet + `
resource "opennebula_cluster" "example" {
	name = "test"

	tags = {
	  environment = "example"
	}
  }
`

var testAccClusterConfigBasic = testAccClusterVnet + `
resource "opennebula_cluster" "example" {
	name = "test"

	hosts = [
	  0
	]
	datastores = [
	  2,
	]
	virtual_networks = [
		opennebula_virtual_network.test.id,
	]

	tags = {
	  environment = "example"
	}
  }
`

var testAccClusterConfigUpdate = testAccClusterVnet + `
resource "opennebula_cluster" "example" {
	name = "test"

	hosts = [
	  0
	]
	datastores = [
	  1,
	]
	virtual_networks = [
		opennebula_virtual_network.test2.id,
	]

	tags = {
	  environment = "updated"
	}
  }
`
