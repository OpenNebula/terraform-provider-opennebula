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
					resource.TestCheckResourceAttr("opennebula_host.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test2", "name", "test2"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.environment", "example"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "name", "test2"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "tags.environment", "example"),
				),
			},
			{
				Config: testAccClusterConfigUpdateMembership,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_host.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test2", "name", "test2"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.environment", "example"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "name", "test2"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "tags.environment", "example"),
				),
			},
			{
				Config: testAccClusterConfigUpdateMembership2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_host.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.environment", "example"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "name", "test2"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "tags.environment", "example"),
				),
			},
			{
				Config: testAccClusterConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_host.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_network.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "name", "test"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.environment", "example"),
					resource.TestCheckResourceAttr("opennebula_cluster.test", "tags.environment2", "example2"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "name", "test2"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_cluster.test2", "tags.environment", "updated2"),
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

var testAccClusterConfigBasic = `
resource "opennebula_host" "test" {
	name       = "test"
	type       = "custom"

	custom {
		virtualization = "dummy"
		information = "dummy"
	}
  
	tags = {
	  test = "test"
	  environment = "example"
	}
  }

resource "opennebula_virtual_network" "test" {
	name            = "test"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500

	security_groups = [0]
	cluster_ids = [opennebula_cluster.test.id, opennebula_cluster.test2.id]
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

	cluster_ids = [opennebula_cluster.test2.id]
	permissions = "660"
	group = "users"

	lifecycle {
	  ignore_changes = [ar, hold_ips]
	}
}

resource "opennebula_cluster" "test" {
	name = "test"

	tags = {
	  environment = "example"
	}

	lifecycle {
		ignore_changes = [hosts, datastores, virtual_networks]
	}
  }

  resource "opennebula_cluster" "test2" {
	name = "test2"

	tags = {
	  environment = "example"
	}

	lifecycle {
		ignore_changes = [hosts, datastores, virtual_networks]
	}
  }
`

var testAccClusterConfigUpdateMembership = `
resource "opennebula_host" "test" {
	name       = "test"
	type       = "custom"
	cluster_id = opennebula_cluster.test2.id

	custom {
		virtualization = "dummy"
		information = "dummy"
	}

	tags = {
	  test = "test"
	  environment = "example"
	}
  }

resource "opennebula_virtual_network" "test" {
	name            = "test"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500

	security_groups = [0]
	cluster_ids = [opennebula_cluster.test2.id]
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

	cluster_ids = [opennebula_cluster.test.id, opennebula_cluster.test2.id]
	permissions = "660"
	group = "users"

	lifecycle {
		ignore_changes = [ar, hold_ips]
	}
}

resource "opennebula_cluster" "test" {
	name = "test"

	tags = {
	  environment = "example"
	}

	lifecycle {
		ignore_changes = [hosts, datastores, virtual_networks]
	}
  }

  resource "opennebula_cluster" "test2" {
	name = "test2"

	tags = {
	  environment = "example"
	}

	lifecycle {
		ignore_changes = [hosts, datastores, virtual_networks]
	}
  }
`

var testAccClusterConfigUpdateMembership2 = `
resource "opennebula_host" "test" {
	name       = "test"
	type       = "custom"
	cluster_id = opennebula_cluster.test.id

	custom {
		virtualization = "dummy"
		information = "dummy"
	}

	tags = {
	  test = "test"
	  environment = "example"
	}
  }
  
resource "opennebula_virtual_network" "test" {
	name            = "test"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500

	security_groups = [0]
	cluster_ids = [opennebula_cluster.test2.id]
	permissions = "660"
	group = "users"

	lifecycle {
		ignore_changes = [ar, hold_ips]
	}
}

resource "opennebula_cluster" "test" {
	name = "test"

	tags = {
		environment = "example"
	}

	lifecycle {
		ignore_changes = [hosts, datastores, virtual_networks]
	}
  }

  resource "opennebula_cluster" "test2" {
	name = "test2"

	tags = {
		environment = "example"
	}

	lifecycle {
		ignore_changes = [hosts, datastores, virtual_networks]
	}
  }
`

var testAccClusterConfigUpdate = `
resource "opennebula_host" "test" {
	name       = "test"
	type       = "custom"
	cluster_id = opennebula_cluster.test.id

	custom {
		virtualization = "dummy"
		information = "dummy"
	}

	tags = {
	  test = "test"
	  environment = "example"
	}
  }
  
resource "opennebula_virtual_network" "test" {
	name            = "test"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500

	security_groups = [0]
	cluster_ids = [opennebula_cluster.test2.id]
	permissions = "660"
	group = "users"

	lifecycle {
	  ignore_changes = [ar, hold_ips]
	}
}

resource "opennebula_cluster" "test" {
	name = "test"

	tags = {
	  environment = "example"
	  environment2 = "example2"
	}


	lifecycle {
		ignore_changes = [hosts, datastores, virtual_networks]
	}
  }

  resource "opennebula_cluster" "test2" {
	name = "test2"

	tags = {
	  environment = "updated2"
	}

	lifecycle {
		ignore_changes = [hosts, datastores, virtual_networks]
	}
  }
`
