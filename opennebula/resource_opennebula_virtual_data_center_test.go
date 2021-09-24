package opennebula

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

func TestAccVirtualDataCenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVirtualDataCenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualDataCenterConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_data_center.vdc", "name", "terravdc"),
					testAccCheckVirtualDataCenterGroups([]int{0}),
					testAccCheckVirtualDataCenterZones(0,
						map[string]interface{}{
							"id":            0,
							"cluster_ids":   []int{0},
							"host_ids":      []int{0},
							"datastore_ids": []int{0, 1, 2},
							"vnet_ids":      []int{},
						},
					),
				),
			},
			{
				Config: testAccVirtualDataCenterConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_data_center.vdc", "name", "terravdc2"),
					testAccCheckVirtualDataCenterGroups([]int{0, 1}),
					testAccCheckVirtualDataCenterZones(0,
						map[string]interface{}{
							"id":            0,
							"cluster_ids":   []int{0},
							"host_ids":      []int{0},
							"datastore_ids": []int{0, 2},
							"vnet_ids":      []int{},
						},
					),
				),
			},
		},
	})
}

func testAccCheckVirtualDataCenterDestroy(s *terraform.State) error {
	controller := testAccProvider.Meta().(*goca.Controller)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opennebula_virtual_data_center" {
			continue
		}
		vdcID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
		vdcc := controller.VDC(int(vdcID))
		// Get Virtual Data Center Info
		vdc, _ := vdcc.Info(false)
		if vdc != nil {
			return fmt.Errorf("Expected VDC %s to have been destroyed", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckVirtualDataCenterGroups(slice []int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "opennebula_virtual_data_center" {
				continue
			}
			vdcID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vdcc := controller.VDC(int(vdcID))
			// Get Virtual Data Center Info
			vdc, _ := vdcc.Info(false)
			if vdc == nil {
				return fmt.Errorf("Expected VDC %s to exist when checking permissions", rs.Primary.ID)
			}
			if !reflect.DeepEqual(vdc.Groups.ID, slice) {
				return fmt.Errorf("VDC (%+v) Groups are not the expected ones, got %+v instead of %+v", vdc, vdc.Groups.ID, slice)
			}
		}
		return nil
	}
}

func testAccCheckVirtualDataCenterZones(zoneidx int, expected map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		controller := testAccProvider.Meta().(*goca.Controller)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "opennebula_virtual_data_center" {
				continue
			}
			vdcID, _ := strconv.ParseUint(rs.Primary.ID, 10, 64)
			vdcc := controller.VDC(int(vdcID))
			// Get Virtual Data Center Info
			vdc, _ := vdcc.Info(false)
			if vdc == nil {
				return fmt.Errorf("Expected VDC %s to exist when checking permissions", rs.Primary.ID)
			}
			zones := flattenZones(vdc)
			for i, zone := range zones {
				if i != zoneidx {
					continue
				}

				for k, _ := range zone {
					// compare id
					if k == "id" {
						if zone[k].(int) != expected[k].(int) {
							return fmt.Errorf("VDC (%s) Zone resources ID lists differ, got %+v instead of %+v", rs.Primary.ID, zone, expected)
						}
						continue
					}
					// compare slice of IDs
					ids := zone[k].([]int)
					if len(ids) != len(expected[k].([]int)) {
						return fmt.Errorf("VDC (%s) Zone resources ID lists differ, got %+v instead of %+v", rs.Primary.ID, zone, expected)
					}
					expectedIDs := expected[k].([]int)
					for i := range ids {
						if ids[i] != expectedIDs[i] {
							return fmt.Errorf("VDC (%s) Zone resources ID lists differ, got %+v instead of %+v", rs.Primary.ID, zone, expected)
						}
					}

				}

			}
		}
		return nil
	}
}

var testAccVirtualDataCenterConfigBasic = `
data opennebula_group "admin" {
  name = "oneadmin"
}

resource "opennebula_virtual_data_center" "vdc" {
  name = "terravdc"
  group_ids = ["${data.opennebula_group.admin.id}"]
  zones {
    id = 0
    host_ids = [0]
    datastore_ids = [0, 1, 2]
    cluster_ids = [0]
  }
}
`

var testAccVirtualDataCenterConfigUpdate = `
data opennebula_group "admin" {
  name = "oneadmin"
}

data opennebula_group "user" {
  name = "users"
}

resource "opennebula_virtual_data_center" "vdc" {
  name = "terravdc2"
  group_ids = ["${data.opennebula_group.admin.id}", "${data.opennebula_group.user.id}"]
  zones {
    id = 0
    host_ids = [0]
    datastore_ids = [0, 2]
    cluster_ids = [0]
  }
}
`
