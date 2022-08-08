package opennebula

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDiskUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolatileDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
				),
			},
			{
				Config: testAccPersistentDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "target", "vdb"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "size", "8"),
				),
			},
			{
				Config: testAccPersistentDiskSizeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "target", "vdb"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "size", "16"),
				),
			},
			{
				Config: testAccSwitchNonPersistentDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "target", "vdc"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "size", "16"),
				),
			},
			{
				Config: testAccNonPersistentDiskTargetUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "target", "vdd"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "size", "16"),
				),
			},
			{
				Config: testAccNonPersistentDiskSizeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "target", "vdd"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "size", "32"),
				),
			},
			{
				Config: testAccNonPersistentDiskAttachedTwice,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "target", "vdd"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "size", "32"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk2", "target", "vde"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk2", "size", "64"),
				),
			},
			{
				Config: testAccNonPersistentDiskSizeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "target", "vdd"),
					resource.TestCheckResourceAttr("opennebula_disk.testdisk1", "size", "32"),
				),
			},
			{
				Config: testAccDiskDetached,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"), // should be removed manually from state
				),
			},
		},
	})
}

var testAccVMConfigBasic = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine"
  group       = "oneadmin"
  permissions = "642"
  memory = 128
  cpu = 0.1
  description = "VM created for provider acceptance tests"

  context = {
	TESTVAR = "TEST"
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

  sched_requirements = "FREE_CPU > 50"

  timeout = 5
}
`

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

var testAccVolatileDisk = testDiskImageResources + `

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

var testAccPersistentDisk = testDiskImageResources + `

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

	lifecycle {
		ignore_changes = [
		  disk,
		]
	}
}

resource "opennebula_disk" "testdisk1" {
	vm_id = opennebula_virtual_machine.test.id
	image_id = opennebula_image.image2.id
	target = "vdb"
}
`

var testAccPersistentDiskSizeUpdate = testDiskImageResources + `

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

	lifecycle {
		ignore_changes = [
		  disk,
		]
	}
}

resource "opennebula_disk" "testdisk1" {
	vm_id = opennebula_virtual_machine.test.id
	image_id = opennebula_image.image2.id
	target = "vdb"
	size = "16"
}
`

var testAccSwitchNonPersistentDisk = testDiskImageResources + `

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

	  lifecycle {
		ignore_changes = [
		  disk,
		]
	}
}

resource "opennebula_disk" "testdisk1" {
	vm_id = opennebula_virtual_machine.test.id
	image_id = opennebula_image.image1.id
	target = "vdc"
}
`

var testAccNonPersistentDiskTargetUpdate = testDiskImageResources + `

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

	  lifecycle {
		ignore_changes = [
		  disk,
		]
	}
}


resource "opennebula_disk" "testdisk1" {
	vm_id = opennebula_virtual_machine.test.id
	image_id = opennebula_image.image1.id
	target = "vdd"
}
`

var testAccNonPersistentDiskSizeUpdate = testDiskImageResources + `

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

	  lifecycle {
		ignore_changes = [
		  disk,
		]
	}
}

resource "opennebula_disk" "testdisk1" {
	vm_id = opennebula_virtual_machine.test.id
	image_id = opennebula_image.image1.id
	target = "vdd"
	size = 32
}
`

var testAccNonPersistentDiskAttachedTwice = testDiskImageResources + `

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

	  lifecycle {
		ignore_changes = [
		  disk,
		]
	}
}

resource "opennebula_disk" "testdisk1" {
	vm_id = opennebula_virtual_machine.test.id
	image_id = opennebula_image.image1.id
	target = "vdd"
	size = 32
}

resource "opennebula_disk" "testdisk2" {
	vm_id = opennebula_virtual_machine.test.id
	image_id = opennebula_image.image1.id
	target = "vde"
	size = 64
}
`

var testAccDiskDetached = testDiskImageResources + `

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

	disk {
	  volatile_type = "swap"
	  size          = 16
	  target        = "vda"
	}

	lifecycle {
	  ignore_changes = [
	    disk,
	  ]
	}
}
`
