package opennebula

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVirtualMachineNICUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineTemplateConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "0"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigNIC,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.131"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigNICsSameVNet,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.1.computed_ip", "172.16.100.132"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigNICUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.1.computed_ip", "172.16.100.111"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigNICIPUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.1.computed_ip", "172.16.100.112"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigMultipleNICs,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "keep_nic_order", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.112"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.1.computed_ip", "172.16.100.132"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.2.computed_ip", "172.16.100.113"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.3.computed_ip", "172.16.100.133"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigMultipleNICsOrderedUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "keep_nic_order", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.112"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.1.computed_ip", "172.16.100.134"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.2.computed_ip", "172.16.100.113"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.3.computed_ip", "172.16.100.133"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigNICDetached,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "keep_nic_order", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "0"),
				),
			},
		},
	})
}

func TestAccVirtualMachineTemplateNIC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVMTemplateNICResource,
			},
			{
				Config: testAccVirtualMachineTemplateNIC,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_model", "virtio"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_virtio_queues", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "0"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateNICAdd,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_model", "virtio"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_virtio_queues", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.111"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_model", "virtio"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_virtio_queues", "2"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateNIC,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_model", "virtio"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_nic.0.computed_virtio_queues", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "0"),
				),
			},
			{
				Config: testAccVMTemplateNICResource,
			},
		},
	})
}

// reproduce #423 problem: index out of range when reordering the list of NICs to attach
func TestAccVirtualMachineAddNIC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineOneNIC,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
				),
			},
			{
				Config: testAccVirtualMachineTwoNICs,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "2"),
				),
			},
			{
				Config: testAccVirtualMachineNoNics,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "0"),
				),
			},
		},
	})
}

var testNICVNetResources = `

resource "opennebula_security_group" "mysecgroup" {
	name        = "secgroup"

	rule {
	  protocol  = "ALL"
	  rule_type = "OUTBOUND"
	}
	rule {
	  protocol  = "TCP"
	  rule_type = "INBOUND"
	  range     = "80"
	}
	rule {
	  protocol  = "ICMP"
	  rule_type = "INBOUND"
	}
  }

resource "opennebula_virtual_network" "network1" {
	name = "test-net1"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500
	ar {
	  ar_type = "IP4"
	  size    = 12
	  ip4     = "172.16.100.130"
	}
	permissions = "642"
	group = "oneadmin"
	security_groups = [0]
	cluster_ids = [0]

	lifecycle {
		ignore_changes = [clusters]
	}

  }

  resource "opennebula_virtual_network" "network2" {
	name = "test-net2"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500
	ar {
	  ar_type = "IP4"
	  size    = 16
	  ip4     = "172.16.100.110"
	}
	permissions = "642"
	group = "oneadmin"
	security_groups = [0]
	cluster_ids = [0]

	lifecycle {
		ignore_changes = [clusters]
	}
  }
`

var testAccVirtualMachineTemplateConfigNIC = testNICVNetResources + `

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
		security_groups = [opennebula_security_group.mysecgroup.id]
	}

	timeout = 5
}
`

var testAccVirtualMachineTemplateConfigNICsSameVNet = testNICVNetResources + `

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
		security_groups = [opennebula_security_group.mysecgroup.id]
	}
	nic {
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.132"
	}

	timeout = 5
}
`

var testAccVirtualMachineTemplateConfigNICUpdate = testNICVNetResources + `

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
		security_groups = [opennebula_security_group.mysecgroup.id]
	  }
	  nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "172.16.100.111"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigNICIPUpdate = testNICVNetResources + `

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
		security_groups = [opennebula_security_group.mysecgroup.id]
	  }
	  nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "172.16.100.112"
	  }


	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigMultipleNICs = testNICVNetResources + `

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

	  keep_nic_order = false
	  nic {
		network_id = opennebula_virtual_network.network2.id
		ip         = "172.16.100.112"
	  }
	  nic {
		network_id = opennebula_virtual_network.network1.id
		ip         = "172.16.100.132"
	  }
	  nic {
		network_id = opennebula_virtual_network.network2.id
		ip         = "172.16.100.113"
	  }
	  nic {
		network_id = opennebula_virtual_network.network1.id
		ip         = "172.16.100.133"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigMultipleNICsOrderedUpdate = testNICVNetResources + `

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

	  keep_nic_order = true
	  nic {
		network_id = opennebula_virtual_network.network2.id
		ip         = "172.16.100.112"
	  }
	  nic {
		network_id = opennebula_virtual_network.network1.id
		ip         = "172.16.100.134"
	  }
	  nic {
		network_id = opennebula_virtual_network.network2.id
		ip         = "172.16.100.113"
	  }
	  nic {
		network_id = opennebula_virtual_network.network1.id
		ip         = "172.16.100.133"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigNICDetached = testNICVNetResources + `

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

	keep_nic_order = false

	timeout = 5
}
`

var testAccVMTemplateNICResource = testNICVNetResources + `

resource "opennebula_template" "template" {
    name        = "test-template"
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

    nic {
	  network_id = opennebula_virtual_network.network1.id
	  ip = "172.16.100.131"
	  model = "virtio"
	  virtio_queues = "2"
    }

    os {
      arch = "x86_64"
      boot = ""
    }

}
`

var testAccVirtualMachineTemplateNIC = testAccVMTemplateNICResource + `

resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"

	template_id = opennebula_template.template.id

	timeout = 5
}
`

var testAccVirtualMachineTemplateNICAdd = testAccVMTemplateNICResource + `

resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"

	template_id = opennebula_template.template.id

	nic {
	  network_id = opennebula_virtual_network.network2.id
	  ip = "172.16.100.111"
	  model = "virtio"
	  virtio_queues = "2"
	}

	timeout = 5
}
`

var testAccVirtualMachineOneNIC = testNICVNetResources + `
resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"
	cpu         = 1
	vcpu        = 2
	memory      = 128

	context = {
	  SET_HOSTNAME = "$NAME"
	  NETWORK      = "YES"
	}

	nic {
	  model           = "virtio"
	  network_id      = opennebula_virtual_network.network1.id
	}
  }
`

var testAccVirtualMachineTwoNICs = testNICVNetResources + `
resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"
	cpu         = 1
	vcpu        = 2
	memory      = 128
  
	context = {
	  SET_HOSTNAME = "$NAME"
	  NETWORK      = "YES"
	}
  
	nic {
		model           = "virtio"
		network_id      = opennebula_virtual_network.network1.id
	}
	nic {
		model           = "virtio"
		network_id      = opennebula_virtual_network.network2.id
	}
  }
`

var testAccVirtualMachineNoNics = testNICVNetResources + `
resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"
	cpu         = 1
	vcpu        = 2
	memory      = 128

	context = {
	  SET_HOSTNAME = "$NAME"
	  NETWORK      = "YES"
	}
  }
`
