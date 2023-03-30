package opennebula

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVirtualMachineDiskUpdate(t *testing.T) {
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
				),
			},
			{
				Config: testAccVirtualMachineVolatileDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
				),
			},
			{
				Config: testAccVirtualMachinePersistentDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_target", "vdb"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_size", "8"),
				),
			},
			{
				Config: testAccVirtualMachinePersistentDiskSizeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_target", "vdb"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_size", "16"),
				),
			},
			{
				Config: testAccVirtualMachineSwitchNonPersistentDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_target", "vdc"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_size", "16"),
				),
			},
			{
				Config: testAccVirtualMachineNonPersistentDiskTargetUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_target", "vdd"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_size", "16"),
				),
			},
			{
				Config: testAccVirtualMachineNonPersistentDiskSizeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_target", "vdd"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_size", "32"),
				),
			},
			{
				Config: testAccVirtualMachineNonPersistentDiskAttachedTwice,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_target", "vdd"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_size", "32"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.2.computed_target", "vde"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.2.computed_size", "64"),
				),
			},
			{
				Config: testAccVirtualMachineNonPersistentDiskSizeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_target", "vdd"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.1.computed_size", "32"),
				),
			},
			{
				Config: testAccVirtualMachineDiskDetached,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
				),
			},
		},
	})
}

var testDiskImageResources = `
resource "opennebula_image" "image1" {
	name             = "image1"
	type             = "DATABLOCK"
	size             = "16"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
  }
  
  resource "opennebula_image" "image2" {
	name             = "image2"
	type             = "DATABLOCK"
	size             = "8"
	datastore_id     = 1
	persistent       = true
	permissions      = "660"
  }
`

var testAccVirtualMachineVolatileDisk = testDiskImageResources + `

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

	disk {
		volatile_type = "swap"
		size          = 16
		target        = "vda"
	}

	timeout = 5
}
`

var testAccVirtualMachinePersistentDisk = testDiskImageResources + `

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

	disk {
		volatile_type = "swap"
		size          = 16
		target        = "vda"
	}
	disk {
		image_id = opennebula_image.image2.id
		target = "vdb"
	}

	timeout = 5
}
`

var testAccVirtualMachinePersistentDiskSizeUpdate = testDiskImageResources + `

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

	disk {
		volatile_type = "swap"
		size          = 16
		target        = "vda"
	}
	disk {
		image_id = opennebula_image.image2.id
		target = "vdb"
		size = "16"
	}

	timeout = 5
}
`

var testAccVirtualMachineSwitchNonPersistentDisk = testDiskImageResources + `

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

	  disk {
		volatile_type = "swap"
		size          = 16
		target        = "vda"
	  }
	  disk {
		  image_id = opennebula_image.image1.id
		  target = "vdc"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineNonPersistentDiskTargetUpdate = testDiskImageResources + `

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

	  disk {
		volatile_type = "swap"
		size          = 16
		target        = "vda"
	  }
	  disk {
		  image_id = opennebula_image.image1.id
		  target = "vdd"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineNonPersistentDiskSizeUpdate = testDiskImageResources + `

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

	  disk {
		volatile_type = "swap"
		size          = 16
		target        = "vda"
	  }
	  disk {
		  image_id = opennebula_image.image1.id
		  target = "vdd"
          size = 32
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineNonPersistentDiskAttachedTwice = testDiskImageResources + `

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

	  disk {
		volatile_type = "swap"
		size          = 16
		target        = "vda"
	  }
	  disk {
		  image_id = opennebula_image.image1.id
		  target = "vdd"
          size = 32
	  }
	  disk {
		image_id = opennebula_image.image1.id
		target = "vde"
		size = 64
	}
	
	  timeout = 5
}
`

var testAccVirtualMachineDiskDetached = testDiskImageResources + `

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
}
`
