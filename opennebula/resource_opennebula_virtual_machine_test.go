package opennebula

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.NETWORK", "YES"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.TESTVAR", "TEST"),
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "sched_requirements", "FREE_CPU > 50"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "description", "VM created for provider acceptance tests"),
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.%", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.NETWORK", "YES"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.TESTVAR", "TEST"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.TESTVAR2", "TEST2"),
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "sched_requirements", "FREE_CPU > 50"),
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
			{
				Config: testAccVirtualMachineContextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine-renamed"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "196"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.NETWORK", "YES"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.TESTVAR", "UPDATE"),
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "sched_requirements", "FREE_CPU > 50"),
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
			{
				Config: testAccVirtualMachineUserTemplateUpdate1,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine-renamed"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "196"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.NETWORK", "YES"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.TESTVAR", "UPDATE"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "sched_requirements", "CLUSTER_ID!=\"123\""),
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
			{
				Config: testAccVirtualMachineUserTemplateUpdate2,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine-renamed"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "196"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.NETWORK", "YES"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.TESTVAR", "UPDATE"),
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.version", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "description", "This is an acceptance test VM"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "sched_requirements", "CLUSTER_ID!=\"123\""),
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
			{
				Config: testAccVirtualMachineUserTemplateUpdate3,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine-renamed"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "196"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.NETWORK", "YES"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "context.TESTVAR", "UPDATE"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "tags.customer", "test2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "description", "VM created for provider acceptance tests"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "sched_requirements", "CLUSTER_ID!=\"123\""),
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

func TestAccVirtualMachineResize(t *testing.T) {
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "128"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.1"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateAddvCPU,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "vcpu", "1"),
				),
			},
			{
				Config: testAccVirtualMachineResizeCpu,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "0.3"),
				),
			},
			{
				Config: testAccVirtualMachineResizevCpu,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "vcpu", "2"),
				),
			},
			{
				Config: testAccVirtualMachineResizeMemory,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "256"),
				),
			},
			{
				Config: testAccVirtualMachineResizePoweroffHard,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
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

func testAccCheckVirtualMachineDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

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
			config := testAccProvider.Meta().(*Configuration)
			controller := config.Controller

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
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

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

  disk {}

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
  description = "VM created for provider acceptance tests"

  context = {
	TESTVAR = "TEST"
	TESTVAR2 = "TEST2"
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

  sched_requirements = "FREE_CPU > 50"

  timeout = 4
}
`

var testAccVirtualMachineContextUpdate = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine-renamed"
  group       = "oneadmin"
  permissions = "660"
  memory = 196
  cpu = 0.2
  description = "VM created for provider acceptance tests"

  context = {
	TESTVAR = "UPDATE"
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

  sched_requirements = "FREE_CPU > 50"

  timeout = 4
}
`

var testAccVirtualMachineUserTemplateUpdate1 = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine-renamed"
  group       = "oneadmin"
  permissions = "660"
  memory = 196
  cpu = 0.2

  context = {
	TESTVAR = "UPDATE"
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
  }

  sched_requirements = "CLUSTER_ID!=\"123\""

  timeout = 4
}
`

var testAccVirtualMachineUserTemplateUpdate2 = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine-renamed"
  group       = "oneadmin"
  permissions = "660"
  memory = 196
  cpu = 0.2
  description = "This is an acceptance test VM"

  context = {
	TESTVAR = "UPDATE"
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
    version = "3"
  }

  sched_requirements = "CLUSTER_ID!=\"123\""

  timeout = 4
}
`

var testAccVirtualMachineUserTemplateUpdate3 = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine-renamed"
  group       = "oneadmin"
  permissions = "660"
  memory = 196
  cpu = 0.2
  description = "VM created for provider acceptance tests"

  context = {
	TESTVAR = "UPDATE"
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
    customer = "test2"
  }

  sched_requirements = "CLUSTER_ID!=\"123\""

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

var testNICVNetResources = `

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

var testAccVirtualMachineTemplateAddvCPU = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine"
  group       = "oneadmin"
  permissions = "642"
  memory = 128
  cpu = 0.1
  vcpu = 1
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

  disk {}

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

var testAccVirtualMachineResizeCpu = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine"
  group       = "oneadmin"
  permissions = "642"
  memory = 128
  cpu = 0.3
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

  disk {}

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

var testAccVirtualMachineResizevCpu = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine"
  group       = "oneadmin"
  permissions = "642"
  memory = 128
  cpu = 0.3
  vcpu = 2
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

  disk {}

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

var testAccVirtualMachineResizeMemory = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine"
  group       = "oneadmin"
  permissions = "642"
  memory = 256
  cpu = 0.3
  vcpu  = 2
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

  disk {}

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

var testAccVirtualMachineResizePoweroffHard = `
resource "opennebula_virtual_machine" "test" {
  name        = "test-virtual_machine"
  group       = "oneadmin"
  permissions = "642"
  memory = 256
  cpu = 0.3
  vcpu  = 2
  hard_shutdown = true
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

  disk {}

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
