package opennebula

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"reflect"
	"strconv"
	"testing"

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
					testAccCheckVirtualDataCenterZones(0, vdcZone{
						ID:           0,
						ClusterIDs:   []int{0},
						HostIDs:      []int{0},
						DatastoreIDs: []int{0, 1, 2},
					}),
				),
			},
			{
				Config: testAccVirtualDataCenterConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("opennebula_virtual_data_center.vdc", "name", "terravdc2"),
					testAccCheckVirtualDataCenterGroups([]int{0, 1}),
					testAccCheckVirtualDataCenterZones(0, vdcZone{
						ID:           0,
						ClusterIDs:   []int{0},
						HostIDs:      []int{0},
						DatastoreIDs: []int{0, 2},
					}),
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
			if !reflect.DeepEqual(vdc.GroupsID, slice) {
				return fmt.Errorf("VDC (%+v) Groups are not the expected ones, got %+v instead of %+v", vdc, vdc.GroupsID, slice)
			}
		}
		return nil
	}
}

func testAccCheckVirtualDataCenterZones(zoneidx int, expected vdcZone) resource.TestCheckFunc {
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
			zones := generateZoneMapFromStructs(vdc)
			for i, zone := range zones {
				if i != zoneidx {
					continue
				}
				if !reflect.DeepEqual(zone, structs.Map(expected)) {
					return fmt.Errorf("VDC (%s) Zones are not the expected ones, got %+v instead of %+v", rs.Primary.ID, zones, expected)
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
