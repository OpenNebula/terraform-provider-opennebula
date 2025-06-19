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
				Config:      testAccVMOneNICOneAliasParentID,
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
				Config: testAccVMDeleteNICAliasBase,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "name", "test-nic-alias-delete"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.0.network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.0.ip", "172.16.100.152"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.0.name", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.1.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.1.ip", "192.168.100.2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.1.name", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.#", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.computed_ip", "172.16.100.132"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.1.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.1.network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.1.computed_ip", "172.16.100.123"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.2.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.2.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.2.computed_ip", "192.168.100.5"),
				),
			},
			{
				Config: testAccVMDeleteOneAlias,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "name", "test-nic-alias-delete"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.computed_ip", "172.16.100.132"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.1.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.1.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.1.computed_ip", "192.168.100.5"),
				),
			},
			{
				Config: testAccVMDeleteFirstAlias,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "name", "test-nic-alias-delete"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.computed_ip", "192.168.100.5"),
				),
			},
			{
				Config: testAccVMDeleteFirstAliasAddNewOne,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "name", "test-nic-alias-delete"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_delete", "nic_alias.0.computed_ip", "192.168.100.8"),
				),
			},
		},
	})
}

func TestAccVirtualMachineUpdateNICAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testNICAliasUpdateBaseResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.ip", "172.16.100.150"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.name", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.ip", "192.168.100.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.name", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "192.168.100.4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "172.16.100.140"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "192.168.100.10"),
				),
			},
			{
				//Update parent NIC on nic_alias 2 (forces recreation) without keeping order
				Config:             testAccVMUpdateNICAliasParent,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "keep_nic_order", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "192.168.100.4"),
					// as we are not keeping order, nic_alias 2 will be recreated and attached at the end of nics list and previous nic 3 will be swapped to pos 2
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "192.168.100.10"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "172.16.100.140"),
				),
			},
			{
				//Fixes previous plan diff reordering the nics in the resource (plan diff should be empty)
				Config: testAccVMUpdateNICAliasParentFixedOrder,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "keep_nic_order", "false"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "192.168.100.4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "192.168.100.10"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "172.16.100.140"),
				),
			},
			{
				//Update network on nic_alias 1 (forces recreation) keeping nic_alias order
				Config: testAccVMUpdateNICAliasNetwork,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "keep_nic_order", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "192.168.100.10"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "172.16.100.140"),
				),
			},
			/*{
			      // Changing the order of already existing NIC aliases in the resource configuration does not affect their actual order.
			      // The provider always orders NICs according to the OpenNebula API response, not the user-defined order.
			      // As a result, reordering aliases in the resource will be ignored unless their parameters are changed.
			      Config: testAccVMUpdateNICAliasOrder,
			      Check: resource.ComposeTestCheckFunc(
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "keep_nic_order", "true"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-0"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net1"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "172.16.100.131"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-0"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net2"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "192.168.100.10"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net1"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "172.16.100.140"),
			       ),
			  },
			  {
			      //change order of aliases, and modify nic_alias1 network and ip
			      Config: testAccVMUpdateNICAliasOrderChangingParameter,
			      Check: resource.ComposeTestCheckFunc(
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "keep_nic_order", "true"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-0"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net2"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "192.168.100.13"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-0"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net2"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "192.168.100.10"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net1"),
			          resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "172.16.100.140"),
			       ) ,
			  },*/
		},
	})
}

// Tests behavior of nic_aliases when parent nics are updated/deleted
func TestAccVirtualMachineUpdateNICAliasParentNIC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testNICAliasUpdateBaseResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.ip", "172.16.100.150"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.name", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.ip", "192.168.100.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.name", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "192.168.100.4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "172.16.100.140"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "192.168.100.10"),
				),
			},
			{
				Config: testNICAliasChangeParentParameterWithKeepOrder,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "keep_nic_order", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.ip", "172.16.100.130"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.name", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.ip", "192.168.100.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.name", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "192.168.100.4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "172.16.100.140"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "192.168.100.10"),
				),
			},
            {
				Config: testNICAliasChangeParentAndDependantNICAliasParameterWithKeepOrder,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "keep_nic_order", "true"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.#", "2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.ip", "172.16.100.134"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.name", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.ip", "192.168.100.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.1.name", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "4"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "192.168.100.7"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "172.16.100.136"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.parent", "test-nic-0"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.3.computed_ip", "192.168.100.10"),
				),
			},
			{
				Config: testAccVMDeleteParentNICMaintainAliases,
				//TODO: expecterror: improve the error regex
				ExpectError: regexp.MustCompile(`.*`),
			},
			{
				Config: testAccVMDeleteParentNICDeleteNicAndAliases,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "name", "test-nic-alias-update"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.ip", "192.168.100.1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic.0.name", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.#", "3"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_network", "test-net1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.0.computed_ip", "172.16.100.147"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.1.computed_ip", "192.168.100.7"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.parent", "test-nic-1"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_network", "test-net2"),
					resource.TestCheckResourceAttr("opennebula_virtual_machine.test_nic_alias_update", "nic_alias.2.computed_ip", "192.168.100.10"),
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

var testAccVMDeleteNICAliasBase = testNICAliasVNetResources + `
resource "opennebula_virtual_machine" "test_nic_alias_delete" {
	name        = "test-nic-alias-delete"
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

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.2"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
        ip = "172.16.100.132"
    }

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
        ip = "172.16.100.123"
    }

    nic_alias {
        parent = "test-nic-1"
        network = opennebula_virtual_network.network2.name
        ip = "192.168.100.5"
    }

	timeout = 5
}
`

var testAccVMDeleteOneAlias = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_delete" {
	name        = "test-nic-alias-delete"
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

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.2"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-0"
        network = opennebula_virtual_network.network1.name
        ip = "172.16.100.132"
    }

    nic_alias {
        parent = "test-nic-1"
        network = opennebula_virtual_network.network2.name
        ip = "192.168.100.5"
    }

	timeout = 5
}
`

var testAccVMDeleteFirstAlias = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_delete" {
	name        = "test-nic-alias-delete"
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

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.2"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network = opennebula_virtual_network.network2.name
        ip = "192.168.100.5"
    }

	timeout = 5
}
`

var testAccVMDeleteFirstAliasAddNewOne = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_delete" {
	name        = "test-nic-alias-delete"
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

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.2"
        name = "test-nic-1"
	}


    nic_alias {
        parent = "test-nic-1"
        network = opennebula_virtual_network.network2.name
        ip = "192.168.100.8"
    }

	timeout = 5
}
`

var testNICAliasUpdateBaseResource = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.4"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.140"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

	timeout = 5
}
`

var testAccVMUpdateNICAliasParent = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.150"
        name = "test-nic-0"
	}

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.4"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.140"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

	timeout = 5
}
`

var testAccVMUpdateNICAliasParentFixedOrder = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.150"
        name = "test-nic-0"
	}

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.4"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.140"
	}

	timeout = 5
}
`

var testAccVMUpdateNICAliasNetwork = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.150"
        name = "test-nic-0"
	}

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.131"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.140"
	}

	timeout = 5
}
`

var testAccVMUpdateNICAliasOrder = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.150"
        name = "test-nic-0"
	}

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.140"
    }

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.131"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

	timeout = 5
}
`

var testAccVMUpdateNICAliasOrderChangingParameter = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.150"
        name = "test-nic-0"
	}

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.13"
    }

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.131"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

	timeout = 5
}
`

var testNICAliasChangeParentParameterWithKeepOrder = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.130"
        name = "test-nic-0"
	}

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.4"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.140"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

	timeout = 5
}
`

var testNICAliasChangeParentAndDependantNICAliasParameterWithKeepOrder = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.134"
        name = "test-nic-0"
	}

    nic {
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.7"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.136"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

	timeout = 5
}
`

var testAccVMDeleteParentNICMaintainAliases = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.7"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.136"
	}

    nic_alias {
        parent = "test-nic-0"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

	timeout = 5
}
`

var testAccVMDeleteParentNICDeleteNicAndAliases = testNICAliasVNetResources + `

resource "opennebula_virtual_machine" "test_nic_alias_update" {
	name        = "test-nic-alias-update"
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
		network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.1"
        name = "test-nic-1"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network1.id
		ip = "172.16.100.147"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.7"
	}

    nic_alias {
        parent = "test-nic-1"
        network_id = opennebula_virtual_network.network2.id
		ip = "192.168.100.10"
	}

	timeout = 5
}
`
