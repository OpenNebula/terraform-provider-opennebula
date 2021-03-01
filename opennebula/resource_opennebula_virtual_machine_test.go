package opennebula

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	ds "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/datastore"
	dskeys "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/datastore/keys"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func TestAccVirtualMachine(t *testing.T) {
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "128"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "timeout", "5"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gname"),
					testAccCheckVirtualMachinePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
				),
			},
			{
				Config: testAccVirtualMachineConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine-renamed"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "196"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.version", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "timeout", "4"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gname"),
					testAccCheckVirtualMachinePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						OwnerA: 0,
						GroupU: 1,
						GroupM: 1,
					}),
				),
			},
		},
	})
}

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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vda"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "8"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigDiskUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vdb"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigDiskTargetUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vdc"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "16"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigDiskSizeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_target", "vdc"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.0.computed_size", "32"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigDiskDetached,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
				),
			},
		},
	})
}

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
				Config: testAccVirtualMachineTemplateConfigNICUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.111"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigNICIPUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.0.computed_ip", "172.16.100.112"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigNICDetached,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "nic.#", "0"),
				),
			},
		},
	})
}

func TestAccVirtualMachinePending(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachinePending,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "virtual_machine_pending"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "128"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "pending", "true"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gname"),
					testAccCheckVirtualMachinePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
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
				Config: testAccVirtualMachineTemplateNICInit,
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
				Config: testAccVirtualMachineTemplateNICInit,
			},
		},
	})
}

func testAccCheckVirtualMachineDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "opennebula_image":
			imgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			imgc := controller.Image(int(imgID))
			// Get Virtual Machine Info
			img, _ := imgc.Info(false)
			if img != nil {
				imgState, _ := img.State()
				if imgState != 6 {
					return fmt.Errorf("Expected image %s to have been destroyed. imgState: %v", rs.Primary.ID, imgState)
				}
			}

		case "opennebula_virtual_machine":
			vmID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vmc := controller.VM(int(vmID))
			// Get Virtual Machine Info
			vm, _ := vmc.Info(false)
			if vm != nil {
				vmState, _, _ := vm.State()
				if vmState != 6 {
					return fmt.Errorf("Expected virtual machine %s to have been destroyed. vmState: %v", rs.Primary.ID, vmState)
				}
			}
		default:
		}

	}

	return nil
}

func testAccSetDSdummy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v := os.Getenv("TF_ACC_VM"); v == "1" {
			controller := testAccProvider.Meta().(*goca.Controller)

			dstpl := ds.NewTemplate()
			dstpl.Add(dskeys.TMMAD, "dummy")
			dstpl.Add(dskeys.DSMAD, "dummy")
			controller.Datastore(0).Update(dstpl.String(), 1)
			controller.Datastore(1).Update(dstpl.String(), 1)
		}
		return nil
	}
}

func testAccCheckVirtualMachinePermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			switch rs.Type {
			case "opennebula_virtual_machine":

				vmID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
				vmc := controller.VM(int(vmID))
				// Get Virtual Machine Info
				vm, err := vmc.Info(false)
				if vm == nil {
					return fmt.Errorf("Expected virtual_machine %s to exist when checking permissions: %s", rs.Primary.ID, err)
				}

				if !reflect.DeepEqual(vm.Permissions, expected) {
					return fmt.Errorf(
						"Permissions for virtual_machine %s were expected to be %s. Instead, they were %s",
						rs.Primary.ID,
						permissionsUnixString(*expected),
						permissionsUnixString(*vm.Permissions),
					)
				}
			default:
			}
		}

		return nil
	}
}

func TestAccVirtualMachineCPUModel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineTemplateConfigCPUModel,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine-renamed"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "196"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpumodel.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpumodel.0.model", "host-passthrough"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "timeout", "5"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.test", "gname"),
					testAccCheckVirtualMachinePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						OwnerA: 0,
						GroupU: 1,
						GroupM: 1,
					}),
				),
			},
		},
	})
}

var testAccVirtualMachineTemplateConfigBasic = `
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

  disk {}

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

var testAccVirtualMachineTemplateConfigCPUModel = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine-renamed"
  group       = "oneadmin"
  permissions = "660"
  memory = 196
  cpu = 0.2

  context = {
    NETWORK  = "YES"
    SET_HOSTNAME = "$NAME"
  }

  graphics {
    type   = "VNC"
    listen = "0.0.0.0"
    keymap = "en-us"
  }

  disk {}

  os {
    arch = "x86_64"
    boot = ""
  }

  cpumodel {
    model = "host-passthrough"
  }

  tags = {
    env = "dev"
    customer = "test"
  }

  timeout = 5
}
`

var testAccVirtualMachineConfigUpdate = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine-renamed"
  group       = "oneadmin"
  permissions = "660"
  memory = 196
  cpu = 0.2

  context = {
    NETWORK  = "YES"
    SET_HOSTNAME = "$NAME"
  }

  graphics {
    type   = "VNC"
    listen = "0.0.0.0"
    keymap = "en-us"
  }

  disk {}

  os {
    arch = "x86_64"
    boot = ""
  }

  tags = {
    env = "dev"
    customer = "test"
    version = "2"
  }
  timeout = 4
}
`

var testAccVirtualMachinePending = `
resource "opennebula_virtual_machine" "test" {
  name        = "virtual_machine_pending"
  group       = "oneadmin"
  permissions = "642"
  memory = 128
  cpu = 0.1
  pending = true

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
}
`

var testAccVirtualMachineTemplateConfigDisk = `

resource "opennebula_image" "img1" {
  name             = "image1"
  type             = "DATABLOCK"
  size             = "16"
  datastore_id     = 1
  persistent       = false
  permissions      = "660"
}

resource "opennebula_image" "img2" {
  name             = "image2"
  type             = "DATABLOCK"
  size             = "8"
  datastore_id     = 1
  persistent       = false
  permissions      = "660"
}

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
		image_id = opennebula_image.img2.id
		target = "vda"
	}

	timeout = 5
}
`

var testAccVirtualMachineTemplateConfigDiskUpdate = `

resource "opennebula_image" "img1" {
	name             = "image1"
	type             = "DATABLOCK"
	size             = "16"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
  }

  resource "opennebula_image" "img2" {
	name             = "image2"
	type             = "DATABLOCK"
	size             = "8"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
  }

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
		  image_id = opennebula_image.img1.id
		  target = "vdb"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigDiskTargetUpdate = `

resource "opennebula_image" "img1" {
	name             = "image1"
	type             = "DATABLOCK"
	size             = "16"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
  }

  resource "opennebula_image" "img2" {
	name             = "image2"
	type             = "DATABLOCK"
	size             = "8"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
  }

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
		  image_id = opennebula_image.img1.id
		  target = "vdc"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigDiskSizeUpdate = `

resource "opennebula_image" "img1" {
	name             = "image1"
	type             = "DATABLOCK"
	size             = "16"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
  }

  resource "opennebula_image" "img2" {
	name             = "image2"
	type             = "DATABLOCK"
	size             = "8"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
  }

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
		  image_id = opennebula_image.img1.id
		  target = "vdc"
          size = 32
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigDiskDetached = `

resource "opennebula_image" "img1" {
  name             = "image1"
  type             = "DATABLOCK"
  size             = "16"
  datastore_id     = 1
  persistent       = false
  permissions      = "660"
}

resource "opennebula_image" "img2" {
  name             = "image2"
  type             = "DATABLOCK"
  size             = "8"
  datastore_id     = 1
  persistent       = false
  permissions      = "660"
}

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

var testAccVirtualMachineTemplateConfigNIC = `

resource "opennebula_virtual_network" "net1" {
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
	clusters = [0]
  }

  resource "opennebula_virtual_network" "net2" {
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
	clusters = [0]
  }


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
		network_id = opennebula_virtual_network.net1.id
		ip = "172.16.100.131"
	}

	timeout = 5
}
`

var testAccVirtualMachineTemplateConfigNICUpdate = `

resource "opennebula_virtual_network" "net1" {
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
	clusters = [0]
  }

  resource "opennebula_virtual_network" "net2" {
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
	clusters = [0]
  }

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
		  network_id = opennebula_virtual_network.net2.id
		  ip = "172.16.100.111"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigNICIPUpdate = `

resource "opennebula_virtual_network" "net1" {
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
	clusters = [0]
  }

  resource "opennebula_virtual_network" "net2" {
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
	clusters = [0]
  }

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
		  network_id = opennebula_virtual_network.net2.id
		  ip = "172.16.100.112"
	  }
	
	  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigNICDetached = `

resource "opennebula_virtual_network" "net1" {
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
	clusters = [0]
  }

  resource "opennebula_virtual_network" "net2" {
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
	clusters = [0]
  }

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
var testAccVirtualMachineTemplateNICInit = `

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
	clusters = [0]
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
	clusters = [0]
}

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

var testAccVirtualMachineTemplateNIC = `

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
	clusters = [0]
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
	clusters = [0]
}

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

resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"

	template_id = opennebula_template.template.id

	timeout = 5
}
`

var testAccVirtualMachineTemplateNICAdd = `

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
	clusters = [0]
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
	clusters = [0]
  }

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
