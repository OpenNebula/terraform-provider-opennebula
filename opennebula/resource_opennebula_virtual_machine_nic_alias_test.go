package opennebula

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVirtualMachineAddNICAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
            {
				Config: testAccVMOneNICOneAliasParentID,
                ExpectError: regexp.MustCompile(`.*`),
			},
			{
				Config: testAccVMOneNICOneAliasWithoutIP,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "name", "test-nic-alias-no-ip"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "nic.#", "1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "nic.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "nic.0.ip", "172.16.100.153"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "nic.0.name", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "nic_alias.#", "1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "nic_alias.0.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "nic_alias.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_no_ip", "nic_alias.0.computed_ip", "172.16.100.100"),
				),
			},
            {
				Config: testAccVMOneNICOneAliasWithIP,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "name", "test-nic-alias-with-ip"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic.#", "1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic.0.ip", "172.16.100.152"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic.0.name", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic_alias.#", "1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic_alias.0.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic_alias.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic_alias.0.computed_ip", "172.16.100.151"),
				),
			},
			{
				Config: testAccVMOneNICTwoAliasesWithIP,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "name", "test-two-nic-aliases-with-ip"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.#", "1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.ip", "172.16.100.150"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.name", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.#", "2"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.computed_network", "test-net2"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.computed_ip", "192.168.100.4"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.computed_network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.computed_ip", "172.16.100.149"),
				),
			},
		},
	})
}


func TestAccVirtualMachineDeleteNICAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
            {
				Config: testAccVMDeleteOneAlias,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "name", "test-nic-alias-with-ip"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic.#", "1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic.0.ip", "172.16.100.152"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic.0.name", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic_alias.#", "0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic_alias.0.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic_alias.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_ip", "nic_alias.0.computed_ip", "172.16.100.151"),
				),
			},
			{
				Config: testAccVMDeleteFirstAlias,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "name", "test-two-nic-aliases-with-ip"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.#", "1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.ip", "172.16.100.150"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.name", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.#", "2"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.computed_network", "test-net2"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.computed_ip", "192.168.100.4"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.computed_network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.computed_ip", "172.16.100.149"),
				),
            },
            {
                Config: testAccVMDeleteFirstAliasAddNewOne,
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "name", "test-two-nic-aliases-with-ip"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.#", "1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.ip", "172.16.100.150"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.0.name", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.#", "2"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.computed_network", "test-net2"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.0.computed_ip", "192.168.100.4"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.parent", "test-nic-0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.computed_network", "test-net1"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.1.computed_ip", "172.16.100.149"),
                ),
            },
            {
                Config: testAccVMDeleteParentNICMaintainAliases,
                //TODO: expecterror: improve the error regex
                ExpectError: regexp.MustCompile(`.*`),
            },
            {
                Config: testAccVMDeleteParentNICDeleteAliases,
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "name", "test-two-nic-aliases-with-ip"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic.#", "0"),
                    resource.TestCheckResourceAttr("opennebula_virtual_machine.test_two_nic_aliases_ip", "nic_alias.#", "0"),
                ),
            },
		},
	})
}


var testNICAliasVNetResources = `

resource "opennebula_virtual_network" "network1" {
	name = "test-net1"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500
	ar {
	  ar_type = "IP4"
	  size    = 56
	  ip4     = "172.16.100.100"
	}
	permissions = "642"
	group = "oneadmin"
	security_groups = [0]
	cluster_ids = [0]
	gateway = "172.16.100.1"
	dns = "8.8.8.8"
  }

  resource "opennebula_virtual_network" "network2" {
	name = "test-net2"
	type            = "dummy"
	bridge          = "onebr"
	mtu             = 1500
	ar {
	  ar_type = "IP4"
	  size    = 16
	  ip4     = "192.168.100.1"
	}
	permissions = "642"
	group = "oneadmin"
	security_groups = [0]
	cluster_ids = [0]
  }
`

var testAccVMOneNICOneAliasParentID = testNICAliasVNetResources + `
resource "opennebula_virtual_machine" "test_nic_alias_parent_id" {
    name        = "test-nic-alias-parent-id"
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
		ip = "172.16.100.155"
	}

    nic_alias {
        network = opennebula_virtual_network.network1.name
        parent_id = "0"
		ip = "172.16.100.154"
	}

	timeout = 5
}
`

var testAccVMOneNICOneAliasWithoutIP = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_no_ip" {
	name        = "test-nic-alias-no-ip"
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
		ip = "172.16.100.153"
        name = "test-nic-0"
	}

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
	}

	timeout = 5
}
`

var testAccVMOneNICOneAliasWithIP = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_ip" {
	name        = "test-nic-alias-with-ip"
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
		ip = "172.16.100.152"
        name = "test-nic-0"
	}

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
		ip = "172.16.100.151"
	}

	timeout = 5
}
`

var testAccVMOneNICTwoAliasesWithIP = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_two_nic_aliases_ip" {
	name        = "test-two-nic-aliases-with-ip"
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
		ip = "172.16.100.150"
        name = "test-nic-0"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.4"
	}

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
        ip = "172.16.100.149"
    }

	timeout = 5
}
`

var testAccVMDeleteOneAlias = testNICAliasVNetResources + testAccVMOneNICOneAliasWithIP + `

resource "opennebula_virtual_machine" "test_nic_alias_ip" {
	name        = "test-nic-alias-with-ip"
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
		ip = "172.16.100.152"
        name = "test-nic-0"
	}

	timeout = 5
}
`

var testAccVMDeleteFirstAlias = testNICAliasVNetResources + testAccVMOneNICTwoAliasesWithIP + `

resource "opennebula_virtual_machine" "test_two_nic_aliases_ip" {
	name        = "test-two-nic-aliases-with-ip"
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
		ip = "172.16.100.150"
        name = "test-nic-0"
	}

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
        ip = "172.16.100.149"
    }

	timeout = 5
}
`

var testAccVMDeleteFirstAliasAddNewOne = testAccVMOneNICTwoAliasesWithIP + `

resource "opennebula_virtual_machine" "test_two_nic_aliases_ip" {
	name        = "test-two-nic-aliases-with-ip"
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
		ip = "172.16.100.150"
        name = "test-nic-0"
	}

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
        ip = "172.16.100.149"
    }

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.4"
	}

	timeout = 5
}
`

var testAccVMDeleteParentNICMaintainAliases = testAccVMOneNICTwoAliasesWithIP + `

resource "opennebula_virtual_machine" "test_two_nic_aliases_ip" {
	name        = "test-two-nic-aliases-with-ip"
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

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
        ip = "172.16.100.149"
    }

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.4"
	}

	timeout = 5
}
`

var testAccVMDeleteParentNICDeleteAliases = testAccVMOneNICTwoAliasesWithIP + `

resource "opennebula_virtual_machine" "test_two_nic_aliases_ip" {
	name        = "test-two-nic-aliases-with-ip"
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
}
`
