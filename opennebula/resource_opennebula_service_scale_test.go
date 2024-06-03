package opennebula

import (
	"context"
	"testing"

	ver "github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// NOTE: OneFlow role template merging is unavailable in OpenNebula releases prior to 6.8.0.
func preCheck(t *testing.T) {
	testAccPreCheck(t)

	if err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil)); err != nil {
		t.Fatal(err)
	}

	config := testAccProvider.Meta().(*Configuration)

	minVersion, _ := ver.NewVersion("6.8.0")

	if config.OneVersion.LessThan(minVersion) {
		t.Skip()
	}
}

func TestAccServiceScale(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { preCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceScaleConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_service.test", "name", "service-scale-test-tf"),
					resource.TestCheckResourceAttr("opennebula_service.test", "roles.0.cardinality", "0"),
					resource.TestCheckResourceAttr("opennebula_service.test", "roles.1.cardinality", "0"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "state"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "template_id"),
				),
			},
			{
				Config: testAccServiceScaleConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_service.test", "name", "service-scale-test-tf"),
					resource.TestCheckResourceAttr("opennebula_service.test", "roles.0.cardinality", "1"),
					resource.TestCheckResourceAttr("opennebula_service.test", "roles.1.cardinality", "1"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "state"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "template_id"),
				),
			},
		},
	})
}

var testAccServiceScaleVMTemplate = `

resource "opennebula_template" "test" {
  name = "service-scale-test-tf"

  cpu    = 1
  vcpu   = 1
  memory = 64

  graphics {
    keymap = "en-us"
    listen = "0.0.0.0"
    type = "VNC"
  }

  os {
    arch = "x86_64"
    boot = ""
  }
}
`

var testAccServiceScaleTemplate = `

resource "opennebula_service_template" "test" {
  name        = "service-scale-test-tf"
  template    = jsonencode({
    TEMPLATE = {
      BODY = {
        name       = "service"
        deployment = "straight"
        roles = [
          {
            name        = "role0"
            cooldown    = 5 # seconds
            vm_template = tonumber(opennebula_template.test.id)
          },
          {
            name        = "role1"
            parents     = ["role0"]
            cooldown    = 5 # seconds
            vm_template = tonumber(opennebula_template.test.id)
          },
        ]
      }
    }
  })
  lifecycle {
    ignore_changes = all
  }
}
`

var testAccServiceScaleConfigBasic = testAccServiceScaleVMTemplate + testAccServiceScaleTemplate + `

resource "opennebula_service" "test" {
  name           = "service-scale-test-tf"
  template_id    = opennebula_service_template.test.id
  extra_template = jsonencode({
    roles = [
      { cardinality = 0 },
      { cardinality = 0 },
    ]
  })
  timeouts {
    create = "2m"
    delete = "2m"
    update = "2m"
  }
}
`

var testAccServiceScaleConfigUpdate = testAccServiceScaleVMTemplate + testAccServiceScaleTemplate + `

resource "opennebula_service" "test" {
  name           = "service-scale-test-tf"
  template_id    = opennebula_service_template.test.id
  extra_template = jsonencode({
    roles = [
      { cardinality = 1 },
      { cardinality = 1 },
    ]
  })
  timeouts {
    create = "2m"
    delete = "2m"
    update = "2m"
  }
}
`
