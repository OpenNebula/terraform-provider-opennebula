//go:build !legacy

package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmkeys "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

func TestAccServiceTemplate(t *testing.T) {
	vm_template_id, _ := setUpServiceTemplateTests()
	tmpl_body := "{\\\"TEMPLATE\\\":{\\\"BODY\\\":{\\\"name\\\":\\\"aa\\\",\\\"deployment\\\":\\\"straight\\\",\\\"roles\\\":[{\\\"name\\\":\\\"master\\\",\\\"cardinality\\\":3,\\\"template_id\\\":"
	tmpl_body = tmpl_body + strconv.Itoa(vm_template_id) + ",\\\"min_vms\\\":2,\\\"type\\\":\\\"vm\\\"}]}}}"
	service_template := testAccServiceTemplateConfigBasic(tmpl_body)
	service_template_update := testAccServiceTemplateConfigUpdate(tmpl_body)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: service_template,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_service_template.test", "name", "tf-tmpl"),
					resource.TestCheckResourceAttr("opennebula_service_template.test", "permissions", "642"),
					resource.TestCheckResourceAttrSet("opennebula_service_template.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_service_template.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_service_template.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_service_template.test", "gname"),
					testAccCheckServiceTemplatePermissions(&shared.Permissions{
						OwnerU: 1,
						OwnerM: 1,
						GroupU: 1,
						OtherM: 1,
					}),
				),
			},
			{
				Config: service_template_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_service_template.test", "name", "tf-tmpl-rename"),
					resource.TestCheckResourceAttr("opennebula_service_template.test", "permissions", "777"),
					resource.TestCheckResourceAttr("opennebula_service_template.test", "uid", "1"),
					resource.TestCheckResourceAttr("opennebula_service_template.test", "gid", "1"),
					resource.TestCheckResourceAttr("opennebula_service_template.test", "uname", "serveradmin"),
					resource.TestCheckResourceAttr("opennebula_service_template.test", "gname", "users"),
					testAccCheckServiceTemplatePermissions(&shared.Permissions{1, 1, 1, 1, 1, 1, 1, 1, 1}),
				),
			},
		},
	})

	tearDownServiceTemplateTests(vm_template_id)
}

func testAccCheckServiceTemplateDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opennebula_service_template" {
			continue
		}
		stID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
		stc := controller.STemplate(int(stID))
		// Get Service Info
		stemplate, _ := stc.Info()
		if stemplate != nil {
			return fmt.Errorf("Expected service template %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckServiceTemplatePermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			stID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			stc := controller.STemplate(int(stID))
			// Get Service
			stemplate, _ := stc.Info()
			if stemplate == nil {
				return fmt.Errorf("Expected service template %s to exist when checking permissions", rs.Primary.ID)
			}

			if !reflect.DeepEqual(stemplate.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for service template %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(*expected),
					permissionsUnixString(*stemplate.Permissions),
				)
			}
		}

		return nil
	}
}

func setUpServiceTemplateTests() (int, error) {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	templateName := "tf-test-template-service"

	// Create template
	tpl := vm.NewTemplate()
	tpl.Add(vmkeys.Name, templateName)
	tpl.CPU(1).Memory(64)

	vmtmpl_id, err := controller.Templates().Create(tpl.String())
	if err != nil {
		return -1, fmt.Errorf("Error creating VM template")
	}

	return vmtmpl_id, nil
}

func tearDownServiceTemplateTests(vm_tmpl int) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	err := controller.Template(vm_tmpl).Delete()
	if err != nil {
		return fmt.Errorf("Error deleting VM template")
	}

	return nil
}

func testAccServiceTemplateConfigBasic(tmpl_body string) string {
	config := "resource \"opennebula_service_template\" \"test\" {\n" +
		"name = \"tf-tmpl\"\n" +
		"template = \"" + tmpl_body + "\"\n" +
		"permissions = \"642\"\n" +
		"uname = \"oneadmin\"\n" +
		"gname = \"oneadmin\"\n" +
		"}\n"

	return config
}

func testAccServiceTemplateConfigUpdate(tmpl_body string) string {
	config := "resource \"opennebula_service_template\" \"test\" {\n" +
		"name = \"tf-tmpl-rename\"\n" +
		"template = \"" + tmpl_body + "\"\n" +
		"permissions = \"777\"\n" +
		"uid = 1\n" +
		"gid = 1\n" +
		"}\n"

	return config
}
