package opennebula

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVirtualMachineDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVirtualMachinesDataSourceInvalidCPU,
				ExpectError: regexp.MustCompile("cpu should be strictly greater than 0"),
			},
			{
				Config:      testAccVirtualMachinesDataSourceInvalidVCPU,
				ExpectError: regexp.MustCompile("vcpu should be strictly greater than 0"),
			},
			{
				Config:      testAccVirtualMachinesDataSourceInvalidMemory,
				ExpectError: regexp.MustCompile("memory should be strictly greater than 0"),
			},
			{
				Config:      testAccVirtualMachinesDataSourceInvalidSort,
				ExpectError: regexp.MustCompile("type \"sort_on\" must be one of: id,name,cpu,vcpu,memory"),
			},
			{
				Config:      testAccVirtualMachinesDataSourceInvalidOrder,
				ExpectError: regexp.MustCompile("type \"order\" must be one of: ASC,DESC"),
			},
			{
				Config:      testAccVirtualMachinesDataSourceNoMatchingVMs,
				ExpectError: regexp.MustCompile("no VMs match the constraints"),
			},
			{
				Config: testAccVirtualMachinesDataSourceBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.basic",
						"virtual_machines.0.name",
						"vm-0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.basic",
						"virtual_machines.0.cpu",
						"0.1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.basic",
						"virtual_machines.0.vcpu",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.basic",
						"virtual_machines.0.memory",
						"128",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.basic",
						"virtual_machines.0.disk.0.volatile_type",
						"swap",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.basic",
						"virtual_machines.0.nic.0.ip",
						"172.20.0.1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.basic",
						"virtual_machines.0.vmgroup.0.vmgroup_id",
						"0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.basic",
						"virtual_machines.0.tags.%",
						"1",
					),
				),
			},
			{
				Config: testAccVirtualMachinesDataSourceBasicSort,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.id_asc",
						"virtual_machines.0.name",
						"vm-1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.id_asc",
						"virtual_machines.1.name",
						"vm-0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.cpu_desc",
						"virtual_machines.0.name",
						"vm-0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.cpu_desc",
						"virtual_machines.1.name",
						"vm-1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.mem_asc",
						"virtual_machines.0.name",
						"vm-0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.mem_asc",
						"virtual_machines.1.name",
						"vm-1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.vcpu_asc",
						"virtual_machines.0.name",
						"vm-1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.vcpu_asc",
						"virtual_machines.1.name",
						"vm-0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.name_desc",
						"virtual_machines.0.name",
						"vm-0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.name_desc",
						"virtual_machines.1.name",
						"vm-1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.id_desc_cpu",
						"virtual_machines.0.name",
						"vm-0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.id_desc_cpu",
						"virtual_machines.0.cpu",
						"0.1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.mem_asc_vcpu",
						"virtual_machines.0.name",
						"vm-1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.mem_asc_vcpu",
						"virtual_machines.0.vcpu",
						"1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.static_name",
						"virtual_machines.0.name",
						"vm-0",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.static_mem",
						"virtual_machines.0.name",
						"vm-1",
					),
					resource.TestCheckResourceAttr(
						"data.opennebula_virtual_machines.static_tags",
						"virtual_machines.0.name",
						"vm-0",
					),
				),
			},
		},
	})
}

var testAccVirtualMachinesDataSourceInvalidCPU = `
data "opennebula_virtual_machines" "test" {
  name_regex = "test.*"
  sort_on    = "id"
  order      = "ASC"
  cpu = 0
}
`
var testAccVirtualMachinesDataSourceInvalidVCPU = `
data "opennebula_virtual_machines" "test" {
  name_regex = "test.*"
  sort_on    = "id"
  order      = "ASC"
  vcpu = 0
}
`
var testAccVirtualMachinesDataSourceInvalidMemory = `
data "opennebula_virtual_machines" "test" {
  name_regex = "test.*"
  sort_on    = "id"
  order      = "ASC"
  memory = 0
}
`
var testAccVirtualMachinesDataSourceInvalidSort = `
data "opennebula_virtual_machines" "test" {
  name_regex = "test.*"
  sort_on    = "unsupported_field"
  order      = "ASC"
}
`
var testAccVirtualMachinesDataSourceInvalidOrder = `
data "opennebula_virtual_machines" "test" {
  name_regex = "test.*"
  sort_on    = "id"
  order      = "unsupported_order"
}
`
var testAccVirtualMachinesDataSourceNoMatchingVMs = `
data "opennebula_virtual_machines" "test" {
  name_regex = "noMatchingVM.*"
}
`
var testAccFirstVirtualMachine = `
resource "opennebula_virtual_machine" "first_vm" {
  name        = "vm-0"
  group       = "oneadmin"
  permissions = "642"
  memory = 128
  cpu = 0.1
  vcpu = 2
  sched_requirements = "CLUSTER_ID!=\"123\""

  disk {
	volatile_type = "swap"
	size          = 16
	target        = "vda"
  }

  nic {
	network_mode_auto = false
	ip = "172.20.0.1"
  }
}
`
var testAccSecondVirtualMachine = `
resource "opennebula_virtual_machine" "second_vm" {	
	name = "vm-1"
	group = "oneadmin"
	permissions = "642"
	memory = 64
	cpu =  0.2
	vcpu = 1
}
`

var testAccVirtualMachinesDataSourceBasic = testAccFirstVirtualMachine + `
data "opennebula_virtual_machines" "basic" {
  name_regex = "vm.*"

  depends_on = [opennebula_virtual_machine.first_vm]
}
`
var testAccVirtualMachinesDataSourceBasicSort = testAccFirstVirtualMachine + testAccSecondVirtualMachine + `
data "opennebula_virtual_machines" "id_asc" {
  name_regex = "vm.*"
  sort_on    = "id"
  order      = "ASC"

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "cpu_desc" {
  name_regex = "vm.*"
  sort_on    = "cpu"
  order      = "DESC"

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "mem_asc" {
  name_regex = "vm.*"
  sort_on    = "memory"
  order      = "ASC"

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "vcpu_asc" {
  name_regex = "vm.*"
  sort_on    = "vcpu"
  order      = "DESC"

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "name_desc" {
  name_regex = "vm.*"
  sort_on    = "name"
  order      = "DESC"

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "id_desc_cpu" {
  name_regex = "vm.*"
  sort_on    = "id"
  cpu = 0.1
  order      = "DESC"

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "mem_asc_vcpu" {
  name_regex = "vm-*"
  sort_on    = "memory"
  vcpu = 1
  order      = "ASC"

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "static_name" {
  name_regex = "vm-0"

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "static_mem" {
  memory = 64

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}

data "opennebula_virtual_machines" "static_tags" {
  tags = {
	sched_requirements = "CLUSTER_ID!=\"123\""
  }

  depends_on = [ 
    opennebula_virtual_machine.first_vm,
  	opennebula_virtual_machine.second_vm 
  ]
}
`
