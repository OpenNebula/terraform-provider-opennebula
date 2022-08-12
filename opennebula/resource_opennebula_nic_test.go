package opennebula

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVMNICUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNIC,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.131"),
				),
			},
			{
				Config: testAccNICsSameVNet,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic1", "ip", "172.16.100.132"),
				),
			},
			{
				Config: testAccNICUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic1", "ip", "172.16.100.111"),
				),
			},
			{
				Config: testAccNICIPUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic1", "ip", "172.16.100.112"),
				),
			},
			{
				Config: testAccNICExportResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"), // should be removed manually from state
					resource.TestCheckResourceAttr("opennebula_nic.testnic1", "ip", "172.16.100.112"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic2", "ip", "172.16.100.131"),
				),
			},
			{
				Config: testAccNICConfigMultiple,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"), // should be removed manually from state
					resource.TestCheckResourceAttr("opennebula_nic.testnic1", "ip", "172.16.100.112"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic2", "ip", "172.16.100.132"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic3", "ip", "172.16.100.113"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic4", "ip", "172.16.100.133"),
				),
			},
			{
				Config: testAccNICConfigMultipleUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "0"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic1", "ip", "172.16.100.112"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic2", "ip", "172.16.100.134"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic3", "ip", "172.16.100.113"),
					resource.TestCheckResourceAttr("opennebula_nic.testnic4", "ip", "172.16.100.133"),
				),
			},
			{
				Config: testAccNICDetached,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "0"),
				),
			},
		},
	})
}

var testAccNIC = testNICVNetResources + `

resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"
	group       = "oneadmin"
	permissions = "642"
	memory = 128
	cpu = 0.1

	context = {
	  NETWORK  = "YES"
	  SET_HOSTNAME = "$NAME"
	}

	graphics {
	  type   = "VNC"
	  listen = "0.0.0.0"
	  keymap = "en-us"
	}

	os {
	  arch = "x86_64"
	  boot = ""
	}

	tags = {
	  env = "prod"
	  customer = "test"
	}

	nic {
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.131"
	}

	timeout = 5

	lifecycle {
		ignore_changes = [
		  nic,
		]
	}
}
`

var testAccNICsSameVNet = testNICVNetResources + `

resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"
	group       = "oneadmin"
	permissions = "642"
	memory = 128
	cpu = 0.1

	context = {
	  NETWORK  = "YES"
	  SET_HOSTNAME = "$NAME"
	}

	graphics {
	  type   = "VNC"
	  listen = "0.0.0.0"
	  keymap = "en-us"
	}

	os {
	  arch = "x86_64"
	  boot = ""
	}

	tags = {
	  env = "prod"
	  customer = "test"
	}

	nic {
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.131"
	}

	lifecycle {
		ignore_changes = [
		  nic,
		]
	}

	timeout = 5
}

resource "opennebula_nic" "testnic1" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network1.id
	ip = "172.16.100.132"
}
`

var testAccNICUpdate = testNICVNetResources + `

  resource "opennebula_virtual_machine" "test" {
	  name        = "test-virtual_machine"
	  group       = "oneadmin"
	  permissions = "642"
	  memory = 128
	  cpu = 0.1
	
	  context = {
		NETWORK  = "YES"
		SET_HOSTNAME = "$NAME"
	  }
	
	  graphics {
		type   = "VNC"
		listen = "0.0.0.0"
		keymap = "en-us"
	  }
	
	  os {
		arch = "x86_64"
		boot = ""
	  }
	
	  tags = {
		env = "prod"
		customer = "test"
	  }

	  nic {
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.131"
	  }
	
	  timeout = 5

	  lifecycle {
		ignore_changes = [
		  nic,
		]
	}
}

resource "opennebula_nic" "testnic1" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network2.id
	ip = "172.16.100.111"
}
`

var testAccNICIPUpdate = testNICVNetResources + `

  resource "opennebula_virtual_machine" "test" {
	  name        = "test-virtual_machine"
	  group       = "oneadmin"
	  permissions = "642"
	  memory = 128
	  cpu = 0.1
	
	  context = {
		NETWORK  = "YES"
		SET_HOSTNAME = "$NAME"
	  }
	
	  graphics {
		type   = "VNC"
		listen = "0.0.0.0"
		keymap = "en-us"
	  }
	
	  os {
		arch = "x86_64"
		boot = ""
	  }
	
	  tags = {
		env = "prod"
		customer = "test"
	  }

	  nic {
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.131"
	  }

	  timeout = 5

	  lifecycle {
		ignore_changes = [
		  nic,
		]
	}
}

resource "opennebula_nic" "testnic1" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network2.id
	ip = "172.16.100.112"
}
`

var testAccNICExportResource = testNICVNetResources + `

  resource "opennebula_virtual_machine" "test" {
	  name        = "test-virtual_machine"
	  group       = "oneadmin"
	  permissions = "642"
	  memory = 128
	  cpu = 0.1
	
	  context = {
		NETWORK  = "YES"
		SET_HOSTNAME = "$NAME"
	  }
	
	  graphics {
		type   = "VNC"
		listen = "0.0.0.0"
		keymap = "en-us"
	  }
	
	  os {
		arch = "x86_64"
		boot = ""
	  }
	
	  tags = {
		env = "prod"
		customer = "test"
	  }

	  timeout = 5

	  lifecycle {
		ignore_changes = [
		  nic,
		]
	}
}

resource "opennebula_nic" "testnic1" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network2.id
	ip = "172.16.100.112"
}

resource "opennebula_nic" "testnic2" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network1.id
	ip = "172.16.100.131"
}
`

var testAccNICConfigMultiple = testNICVNetResources + `

  resource "opennebula_virtual_machine" "test" {
	  name        = "test-virtual_machine"
	  group       = "oneadmin"
	  permissions = "642"
	  memory = 128
	  cpu = 0.1
	
	  context = {
		NETWORK  = "YES"
		SET_HOSTNAME = "$NAME"
	  }
	
	  graphics {
		type   = "VNC"
		listen = "0.0.0.0"
		keymap = "en-us"
	  }
	
	  os {
		arch = "x86_64"
		boot = ""
	  }
	
	  tags = {
		env = "prod"
		customer = "test"
	  }

	  timeout = 5

	  lifecycle {
		ignore_changes = [
		  nic,
		]
	}
}

resource "opennebula_nic" "testnic1" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network2.id
	ip = "172.16.100.112"
}

resource "opennebula_nic" "testnic2" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network1.id
	ip = "172.16.100.132"
}

resource "opennebula_nic" "testnic3" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network2.id
	ip         = "172.16.100.113"
}

resource "opennebula_nic" "testnic4" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network1.id
	ip         = "172.16.100.133"
}
`

var testAccNICConfigMultipleUpdate = testNICVNetResources + `

  resource "opennebula_virtual_machine" "test" {
	  name        = "test-virtual_machine"
	  group       = "oneadmin"
	  permissions = "642"
	  memory = 128
	  cpu = 0.1
	
	  context = {
		NETWORK  = "YES"
		SET_HOSTNAME = "$NAME"
	  }
	
	  graphics {
		type   = "VNC"
		listen = "0.0.0.0"
		keymap = "en-us"
	  }
	
	  os {
		arch = "x86_64"
		boot = ""
	  }
	
	  tags = {
		env = "prod"
		customer = "test"
	  }
	
	  timeout = 5

	  lifecycle {
		ignore_changes = [
		  nic,
		]
	}
}

resource "opennebula_nic" "testnic1" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network2.id
	ip = "172.16.100.112"
}

resource "opennebula_nic" "testnic2" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network1.id
	ip = "172.16.100.134"
}

resource "opennebula_nic" "testnic3" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network2.id
	ip         = "172.16.100.113"
}

resource "opennebula_nic" "testnic4" {
	vm_id = opennebula_virtual_machine.test.id
	network_id = opennebula_virtual_network.network1.id
	ip         = "172.16.100.133"
}
`

var testAccNICDetached = testNICVNetResources + `

resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"
	group       = "oneadmin"
	permissions = "642"
	memory = 128
	cpu = 0.1

	context = {
	  NETWORK  = "YES"
	  SET_HOSTNAME = "$NAME"
	}

	graphics {
	  type   = "VNC"
	  listen = "0.0.0.0"
	  keymap = "en-us"
	}

	os {
	  arch = "x86_64"
	  boot = ""
	}

	tags = {
	  env = "prod"
	  customer = "test"
	}

	timeout = 5

	lifecycle {
		ignore_changes = [
		  nic,
		]
	}
}
`
