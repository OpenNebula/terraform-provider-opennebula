package opennebula

import (
	"fmt"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
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
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_virtual_machine.test", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey1": "testvalue1",
						"elements.testkey2": "testvalue2",
					}),
				),
			},
			{
				Config: testAccVirtualMachineTemplateConfigOs,
				Check: resource.ComposeTestCheckFunc(
					testAccSetDSdummy(),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "memory", "128"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "cpu", "0.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "context.%", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "context.NETWORK", "YES"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "context.TESTVAR", "TEST"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "os.0.firmware", ""),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "os.0.firmware_secure", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "disk.#", "0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "description", "VM created for provider acceptance tests"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.testos", "timeout", "5"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.testos", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.testos", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.testos", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_virtual_machine.testos", "gname"),
					testAccCheckVirtualMachinePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_virtual_machine.testos", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey1": "testvalue1",
						"elements.testkey2": "testvalue2",
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
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
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_virtual_machine.test", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey1": "testvalue_updated",
						"elements.testkey3": "testvalue3",
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "template_section.#", "0"),
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
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
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_virtual_machine.test", "template_section.*", map[string]string{
						"name":              "test_vec_key2",
						"elements.%":        "1",
						"elements.testkey1": "testvalue2",
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
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
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_virtual_machine.test", "template_section.*", map[string]string{
						"name":              "test_vec_key3",
						"elements.%":        "3",
						"elements.testkey1": "testvalue1",
						"elements.testkey2": "testvalue2",
						"elements.testkey3": "testvalue3",
					}),
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

func TestAccVirtualMachineDoneTriggerRecreation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccVirtualMachineDone,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "virtual_machine_done"),
					testAccTerminateVM("virtual_machine_done"),
				),
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					func(state *terraform.State) error {
						if !state.Empty() && len(state.RootModule().Resources) != 0 {
							return fmt.Errorf("expected state to be empty")
						}
						return nil
					},
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

func TestAccVirtualMachineTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineTemplateInstantiate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "vcpu", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "768"),
					resource.TestCheckNoResourceAttr("opennebula_virtual_machine.test", "sched_requirements"),
				),
			},
			{
				Config: testAccVirtualMachineTemplateOverrideKeys,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "name", "test-virtual_machine"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "cpu", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "vcpu", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "memory", "768"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "sched_requirements", "CLUSTER_ID!=\"123\""),
				),
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
			imgID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
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
			vmID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
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

func testAccTerminateVM(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		id, err := controller.VMs().ByName(name)
		if err != nil {
			return err
		}

		err = controller.VM(id).Terminate()
		if err != nil {
			return err
		}
		return waitForVMState(id, vm.Done, time.Minute*5)
	}
}

// waitForVMState waits until the VM with vmId has the desiredState.
// returns an error if timeout is reached.
func waitForVMState(vmId int, desiredState vm.State, timeout time.Duration) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller
	interval := time.NewTicker(5 * time.Second)
	deadline := time.NewTimer(timeout)

	// Ensure ticker and timer are stopped after use
	defer interval.Stop()
	defer deadline.Stop()

	for {
		select {
		case <-interval.C:
			info, err := controller.VM(vmId).Info(false)
			if err != nil {
				return err
			}
			if vm.State(info.StateRaw) == desiredState {
				return nil
			}
		case <-deadline.C:
			return fmt.Errorf("timeout waiting for vm id '%d' to reach desired state %s\n", vmId, desiredState)
		}
	}
}

func testAccCheckVirtualMachinePermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			switch rs.Type {
			case "opennebula_virtual_machine":

				vmID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
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
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test", "disk.#", "0"),
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

  os {
    arch = "x86_64"
    boot = ""
   }

  tags = {
    env = "prod"
    customer = "test"
  }

  sched_requirements = "FREE_CPU > 50"

  template_section {
	name = "test_vec_key"
	elements = {
		testkey1 = "testvalue1"
		testkey2 = "testvalue2"
	}
  }

  timeout = 5
}
`

var testAccVirtualMachineTemplateConfigOs = `
resource "opennebula_virtual_machine" "testos" {
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
    firmware = ""
    firmware_secure = false
  }

  tags = {
    env = "prod"
    customer = "test"
  }

  template_section {
	name = "test_vec_key"
	elements = {
		testkey1 = "testvalue1"
		testkey2 = "testvalue2"
	}
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

  template_section {
	name = "test_vec_key"
	elements = {
		testkey1 = "testvalue_updated"
		testkey3 = "testvalue3"
	}
  }
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

  template_section {
	name = "test_vec_key"
	elements = {
		testkey1 = "testvalue_updated"
		testkey3 = "testvalue3"
	}
  }
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

  template_section {
	name = "test_vec_key2"
	elements = {
		testkey1 = "testvalue2"
	}
  }
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

  template_section {
	name = "test_vec_key3"
	elements = {
		testkey1 = "testvalue1"
		testkey2 = "testvalue2"
		testkey3 = "testvalue3"
	}
  }
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

var testAccVirtualMachineDone = `
resource "opennebula_virtual_machine" "test" {
  name        = "virtual_machine_done"
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

var testAccVirtualMachineTemplate = `
resource "opennebula_template" "template" {
  name = "terratplupdate"
  permissions = "642"
  group = "oneadmin"
  description = "Template created for provider acceptance tests - updated"

  cpu = "1"
  vcpu = "1"
  memory = "768"

  features {
    virtio_scsi_queues = 1
    acpi = "YES"
  }

  context = {
	dns_hostname = "yes"
	network = "YES"
  }

  graphics {
	keymap = "en-us"
	listen = "0.0.0.0"
	type = "VNC"
  }

  os {
	arch = "x86_64"
	boot = ""
  }

  tags = {
    env = "dev"
    customer = "test"
    version = "2"
  }

  sched_requirements = "CLUSTER_ID!=\"123\""

}
`

var testAccVirtualMachineTemplateInstantiate = testAccVirtualMachineTemplate + `
resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"
	group       = "oneadmin"

	template_id = opennebula_template.template.id
}
`

var testAccVirtualMachineTemplateOverrideKeys = testAccVirtualMachineTemplate + `
resource "opennebula_virtual_machine" "test" {
	name        = "test-virtual_machine"
	group       = "oneadmin"

	template_id = opennebula_template.template.id

	sched_requirements = "CLUSTER_ID!=\"123\""
}
`
