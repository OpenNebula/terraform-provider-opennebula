package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	srv_tmpl "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/service_template"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmkeys "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

func TestAccService(t *testing.T) {
	service_template_id, vm_template_id, _ := setUpServiceTests()
	service_template := testAccServiceConfigBasic(service_template_id)
	service_template_update := testAccServiceConfigUpdate(service_template_id)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: service_template,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_service.test", "name", "service-test-tf"),
					resource.TestCheckResourceAttr("opennebula_service.test", "permissions", "642"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "uid"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "gid"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "uname"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "gname"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "state"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "template_id"),
					testAccCheckServicePermissions(&shared.Permissions{
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
					resource.TestCheckResourceAttr("opennebula_service.test", "name", "service-test-tf-renamed"),
					resource.TestCheckResourceAttr("opennebula_service.test", "permissions", "777"),
					resource.TestCheckResourceAttr("opennebula_service.test", "uid", "1"),
					resource.TestCheckResourceAttr("opennebula_service.test", "gid", "1"),
					resource.TestCheckResourceAttr("opennebula_service.test", "uname", "serveradmin"),
					resource.TestCheckResourceAttr("opennebula_service.test", "gname", "users"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "state"),
					resource.TestCheckResourceAttrSet("opennebula_service.test", "template_id"),
					testAccCheckServicePermissions(&shared.Permissions{1, 1, 1, 1, 1, 1, 1, 1, 1}),
				),
			},
		},
	})

	tearDownServiceTests(service_template_id, vm_template_id)
}

func testAccCheckServiceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opennebula_service" {
			continue
		}
		svID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
		sc := controller.Service(int(svID))
		// Get Service Info
		service, _ := sc.Info()
		if service != nil {
			svState := service.Template.Body.StateRaw
			if svState != 5 {
				return fmt.Errorf("Expected service %s to have been destroyed", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckServicePermissions(expected *shared.Permissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Configuration)
		controller := config.Controller

		for _, rs := range s.RootModule().Resources {
			serviceID, _ := strconv.ParseUint(rs.Primary.ID, 10, 0)
			sc := controller.Service(int(serviceID))
			// Get Service
			service, _ := sc.Info()
			if service == nil {
				return fmt.Errorf("Expected service %s to exist when checking permissions", rs.Primary.ID)
			}

			if !reflect.DeepEqual(service.Permissions, expected) {
				return fmt.Errorf(
					"Permissions for service %s were expected to be %s. Instead, they were %s",
					rs.Primary.ID,
					permissionsUnixString(*expected),
					permissionsUnixString(*service.Permissions),
				)
			}
		}

		return nil
	}
}

func setUpServiceTests() (int, int, error) {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	templateName := "tf-test-template-service"

	// Create template
	tpl := vm.NewTemplate()
	tpl.Add(vmkeys.Name, templateName)
	tpl.CPU(1).Memory(64)

	vmtmpl_id, err := controller.Templates().Create(tpl.String())
	if err != nil {
		return -1, -1, fmt.Errorf("Error creating VM template")
	}

	tmpl := srv_tmpl.ServiceTemplate{
		Template: srv_tmpl.Template{
			Body: srv_tmpl.Body{
				Name:       "NewTemplateTest",
				Deployment: "straight",
				Roles: []srv_tmpl.Role{
					{
						Name:        "master",
						Cardinality: 1,
						Type:        "vm",
						TemplateID:  vmtmpl_id,
						MinVMs:      1,
					},
				},
			},
		},
	}

	err = controller.STemplates().Create(&tmpl)
	if err != nil {
		return -1, -1, fmt.Errorf("Error creating service template")
	}

	return tmpl.ID, vmtmpl_id, nil
}

func tearDownServiceTests(sv_tmpl, vm_tmpl int) error {
	config := testAccProvider.Meta().(*Configuration)
	controller := config.Controller

	err := controller.Template(vm_tmpl).Delete()
	if err != nil {
		return fmt.Errorf("Error deleting VM template")
	}

	err = controller.STemplate(sv_tmpl).Delete()
	if err != nil {
		return fmt.Errorf("Error deleting service template")
	}

	return nil
}

func testAccServiceConfigBasic(tmpl_id int) string {
	config := "resource \"opennebula_service\" \"test\" {\n" +
		"	name           = \"service-test-tf\"\n" +
		"	template_id    = " + strconv.Itoa(tmpl_id) + "\n" +
		"	permissions    = \"642\"\n" +
		"	uid            = 0\n" +
		"	gid            = 0\n" +
		"}\n"

	return config
}

func testAccServiceConfigUpdate(tmpl_id int) string {
	config := "resource \"opennebula_service\" \"test\" {\n" +
		"	name           = \"service-test-tf-renamed\"\n" +
		"	template_id    = " + strconv.Itoa(tmpl_id) + "\n" +
		"	permissions    = \"777\"\n" +
		"	uid            = 1\n" +
		"	gid            = 1\n" +
		"}\n"

	return config
}
