package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func TestAccTemplate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template", "name", "terra-tpl"),
					resource.TestCheckResourceAttr("opennebula_template.template", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_template.template", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_template.template", "cpu", "0.5"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.0.target", "vda"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.0.size", "16"),
					resource.TestCheckResourceAttr("opennebula_template.template", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "nic.0.ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.customer", "test"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_template.template", "features.*", map[string]string{
						"virtio_scsi_queues": "1",
						"acpi":               "YES",
					}),
					resource.TestCheckResourceAttr("opennebula_template.template", "sched_requirements", "FREE_CPU > 50"),
					resource.TestCheckResourceAttr("opennebula_template.template", "user_inputs.%", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "user_inputs.BLOG_TITLE", "M|text|Blog Title"),
					resource.TestCheckResourceAttr("opennebula_template.template", "description", "Template created for provider acceptance tests"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gname"),
					testAccCheckTemplatePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						GroupM: 1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_template.template", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey1": "testvalue1",
						"elements.testkey2": "testvalue2",
					}),
				),
			},
			{
				Config: testAccTemplateCPUModel,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template", "name", "terra-tpl-cpumodel"),
					resource.TestCheckResourceAttr("opennebula_template.template", "permissions", "660"),
					resource.TestCheckResourceAttr("opennebula_template.template", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_template.template", "cpu", "0.5"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.0.target", "vda"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.0.size", "16"),
					resource.TestCheckResourceAttr("opennebula_template.template", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "nic.0.ip", "172.16.100.131"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_template.template", "cpumodel.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "cpumodel.0.model", "host-passthrough"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.env", "prod"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_template.template", "sched_requirements", "FREE_CPU > 50"),
					resource.TestCheckResourceAttr("opennebula_template.template", "description", "Template created for provider acceptance tests"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gname"),
					testAccCheckTemplatePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						GroupM: 1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_template.template", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey1": "testvalue1",
						"elements.testkey2": "testvalue2",
					}),
				),
			},
			{
				Config: testAccTemplateConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template", "name", "terratplupdate"),
					resource.TestCheckResourceAttr("opennebula_template.template", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_template.template", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_template.template", "cpu", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.0.target", "vda"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.0.size", "32"),
					resource.TestCheckResourceAttr("opennebula_template.template", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "nic.0.ip", "172.16.100.132"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.%", "3"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.customer", "test"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.version", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_template.template", "features.*", map[string]string{
						"virtio_scsi_queues": "1",
						"acpi":               "YES",
					}),
					resource.TestCheckResourceAttr("opennebula_template.template", "sched_requirements", "CLUSTER_ID!=\"123\""),
					resource.TestCheckResourceAttr("opennebula_template.template", "description", "Template created for provider acceptance tests - updated"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gname"),
					testAccCheckTemplatePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_template.template", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "2",
						"elements.testkey2": "testvalue2",
						"elements.testkey3": "testvalue3",
					}),
				),
			},
			{
				Config: testAccTemplateConfigDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template", "name", "terratplupdate"),
					resource.TestCheckResourceAttr("opennebula_template.template", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_template.template", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_template.template", "cpu", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.keymap", "en-us"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.listen", "0.0.0.0"),
					resource.TestCheckResourceAttr("opennebula_template.template", "graphics.0.type", "VNC"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.0.target", "vda"),
					resource.TestCheckResourceAttr("opennebula_template.template", "disk.0.size", "32"),
					resource.TestCheckResourceAttr("opennebula_template.template", "nic.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "nic.0.ip", "172.16.100.132"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.arch", "x86_64"),
					resource.TestCheckResourceAttr("opennebula_template.template", "os.0.boot", ""),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.%", "2"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.env", "dev"),
					resource.TestCheckResourceAttr("opennebula_template.template", "tags.customer", "test"),
					resource.TestCheckNoResourceAttr("opennebula_template.template", "tags.version"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_template.template", "gname"),
					testAccCheckTemplatePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("opennebula_template.template", "template_section.*", map[string]string{
						"name":              "test_vec_key",
						"elements.%":        "1",
						"elements.testkey3": "testvalue3",
					}),
				),
			},
			{
				Config: testAccTemplateImageDisk,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "name", "terratplimageDisk"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "permissions", "642"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "group", "oneadmin"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "cpu", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "vcpu", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "memory", "768"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "disk.0.image", "imageName"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "disk.0.image_id", "-1"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "disk.#", "1"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "disk.0.target", "vda"),
					resource.TestCheckResourceAttr("opennebula_template.template_disk_test", "disk.0.size", "64"),
					resource.TestCheckResourceAttr("opennebula_image.image", "name", "imageName"),
					resource.TestCheckResourceAttr("opennebula_image.image", "size", "16"),
					resource.TestCheckResourceAttr("opennebula_image.image", "type", "DATABLOCK"),
					resource.TestCheckResourceAttr("opennebula_image.image", "datastore_id", "1"),
					resource.TestCheckResourceAttr("opennebula_image.image", "persistent", "false"),
					resource.TestCheckResourceAttr("opennebula_image.image", "permissions", "660"),
				),
			},
		},
	})
}

func testAccCheckTemplatePermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "opennebula_template" {
				continue
			}
			tID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			tc := controller.Template(int(tID))
			// TODO: fix it after 5.10 release availability
			// Force the "extended" bool to false to keep ONE 5.8 behavior
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			template, _ := tc.Info(false, false)
			if template == nil {
				return fmt.Errorf("Expected template %s to exist when checking permissions", rs.Primary.ID)
			}

			if !reflect.DeepEqual(template.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for template %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(*expected),
					permissionsUnixString(*template.Permissions),
				)
			}
		}

		return nil
	}
}
func testAccCheckTemplateDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller
	var destroy bool

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opennebula_template" {
			continue
		}
		templateID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
		tc := controller.Template(int(templateID))
		// Get Template Info
		template, _ := tc.Info(false, false)
		if template != nil {
			return fmt.Errorf("Expected template %s to have been destroyed", rs.Primary.ID)
		}
		destroy = true
	}

	if !destroy {
		return fmt.Errorf("No resource to be destroyed")
	}

	return nil
}

var testTemplateNICVNetResources = `

resource "opennebula_virtual_network" "network" {
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
  }
`

var testAccTemplateConfigBasic = testTemplateNICVNetResources + `
resource "opennebula_template" "template" {
  name = "terra-tpl"
  permissions = "660"
  group = "oneadmin"
  cpu = "0.5"
  vcpu = "1"
  memory = "512"
  description = "Template created for provider acceptance tests"

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

  disk {
	volatile_type = "swap"
	size          = 16
	target        = "vda"
  }

  nic {
	network_id = opennebula_virtual_network.network.id
	ip = "172.16.100.131"
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

  user_inputs = {
	BLOG_TITLE="M|text|Blog Title"
  }

  template_section {
	name = "test_vec_key"
	elements = {
		testkey1 = "testvalue1"
		testkey2 = "testvalue2"
	}
  }

}
`

var testAccTemplateCPUModel = testTemplateNICVNetResources + `
resource "opennebula_template" "template" {
  name = "terra-tpl-cpumodel"
  permissions = "660"
  group = "oneadmin"
  cpu = "0.5"
  vcpu = "1"
  memory = "512"
  description = "Template created for provider acceptance tests"

  context = {
    dns_hostname = "yes"
    network = "YES"
  }

  graphics {
    keymap = "en-us"
    listen = "0.0.0.0"
    type = "VNC"
  }

  disk {
	volatile_type = "swap"
	size          = 16
	target        = "vda"
  }

  nic {
	network_id = opennebula_virtual_network.network.id
	ip = "172.16.100.131"
  }

  cpumodel {
    model = "host-passthrough"
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

}
`

var testAccTemplateConfigUpdate = testTemplateNICVNetResources + `
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

  disk {
	volatile_type = "swap"
	size          = 32
	target        = "vda"
  }

  nic {
	network_id = opennebula_virtual_network.network.id
	ip = "172.16.100.132"
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

  template_section {
	name = "test_vec_key"
	elements = {
		testkey2 = "testvalue2"
		testkey3 = "testvalue3"
	}
  }

}
`

var testAccTemplateConfigDelete = testTemplateNICVNetResources + `
resource "opennebula_template" "template" {
  name = "terratplupdate"
  permissions = "642"
  group = "oneadmin"

  cpu = "1"
  vcpu = "1"
  memory = "768"

  context = {
	dns_hostname = "yes"
	network = "YES"
  }

  graphics {
	keymap = "en-us"
	listen = "0.0.0.0"
	type = "VNC"
  }

  disk {
	volatile_type = "swap"
	size          = 32
	target        = "vda"
  }

  nic {
	network_id = opennebula_virtual_network.network.id
	ip = "172.16.100.132"
  }

  os {
	arch = "x86_64"
	boot = ""
  }

  tags = {
    env = "dev"
    customer = "test"
  }

  template_section {
	name = "test_vec_key"
	elements = {
		testkey3 = "testvalue3"
	}
  }
}
`

var testAccTemplateImageDisk = `
resource "opennebula_image" "image" {
	name             = "imageName"
	type             = "DATABLOCK"
	size             = "16"
	datastore_id     = 1
	persistent       = false
	permissions      = "660"
  }


resource "opennebula_template" "template_disk_test" {
  name = "terratplimageDisk"
  permissions = "642"
  group = "oneadmin"

  cpu = "1"
  vcpu = "1"
  memory = "768"

  disk {
	image = "imageName"
	size     = 64
	target   = "vda"
  }
}
`
