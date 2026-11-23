package opennebula

import (
	"context"
	"strconv"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
)

func TestVMDiskUUIDGeneration(t *testing.T) {
	for _, managed := range []bool{true, false} {
		diskSchema := diskSchema()
		if managed {
			diskSchema = diskVMSchema()
		}
		d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{"disk": diskSchema}, map[string]interface{}{
			"disk": []interface{}{
				map[string]interface{}{"image_id": 3},
				map[string]interface{}{"image_id": 3},
			},
		})
		tpl := vm.NewTemplate()
		if err := addDisks(d, tpl); err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		if len(tpl.Elements) != 2 {
			t.Fatalf("got %d disk vectors, want 2", len(tpl.Elements))
		}
		for i, element := range tpl.Elements {
			disk := element.(*shared.Disk)
			id, _ := disk.Get("TF_UUID")
			if !managed {
				if id != "" {
					t.Fatal("reusable template disk received a UUID")
				}
				continue
			}
			if _, err := uuid.ParseUUID(id); err != nil || seen[id] {
				t.Fatalf("invalid or duplicate disk UUID %q", id)
			}
			seen[id] = true
			config := d.Get("disk").([]interface{})[i].(map[string]interface{})
			if config["tf_uuid"] != id {
				t.Fatal("disk UUID was not saved in state")
			}
			replacement, err := makeVMDiskVector(config)
			if err != nil {
				t.Fatal(err)
			}
			if next, _ := replacement.Get("TF_UUID"); next != id {
				t.Fatal("managed disk UUID changed during an update")
			}
		}
	}
}

func testVMDisk(id int, tfUUID string, size int, target string) shared.Disk {
	disk := shared.NewDisk()
	disk.Add(shared.DiskID, strconv.Itoa(id))
	disk.Add(shared.ImageID, "3")
	disk.Add(shared.Size, strconv.Itoa(size))
	disk.Add(shared.TargetDisk, target)
	disk.Add(shared.Driver, "qcow2")
	if tfUUID != "" {
		disk.Add("TF_UUID", tfUUID)
	}
	return *disk
}

func TestFlattenVMDiskUUID(t *testing.T) {
	for _, tc := range []struct {
		name, storedUUID, remoteUUID string
		remoteSize, wantCount        int
	}{
		{"external resize", "managed", "managed", 13264, 1},
		{"reused slot", "managed", "replacement", 12240, 0},
		{"missing tag", "managed", "", 12240, 0},
		{"legacy entry cannot claim a tagged disk", "", "other", 12240, 0},
		{"legacy disk", "", "", 12240, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceOpennebulaVirtualMachine().Schema, map[string]interface{}{
				"disk": []interface{}{map[string]interface{}{
					"image_id": 3, "size": 12240, "target": "vda", "driver": "qcow2", "tf_uuid": tc.storedUUID,
				}},
			})
			tpl := vm.NewTemplate()
			disk := testVMDisk(0, tc.remoteUUID, tc.remoteSize, "vda")
			tpl.Elements = append(tpl.Elements, &disk.Vector)
			if err := flattenVMDisk(d, tpl); err != nil {
				t.Fatal(err)
			}
			disks := d.Get("disk").([]interface{})
			if len(disks) != tc.wantCount {
				t.Fatalf("got %d disks, want %d", len(disks), tc.wantCount)
			}
			if len(disks) > 0 {
				got := disks[0].(map[string]interface{})
				if got["size"] != tc.remoteSize || got["computed_size"] != tc.remoteSize || got["tf_uuid"] != tc.remoteUUID {
					t.Fatalf("incorrect refreshed disk: %#v", got)
				}
			}
		})
	}
}

func TestFlattenVMDiskUUIDPreservesOrder(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceOpennebulaVirtualMachine().Schema, map[string]interface{}{
		"disk": []interface{}{
			map[string]interface{}{"image_id": 3, "tf_uuid": "first"},
			map[string]interface{}{"image_id": 3, "tf_uuid": "second"},
		},
	})
	tpl := vm.NewTemplate()
	for _, disk := range []shared.Disk{
		testVMDisk(2, "external", 100, "vdc"),
		testVMDisk(1, "second", 200, "vdb"),
		testVMDisk(0, "first", 300, "vda"),
	} {
		disk := disk
		tpl.Elements = append(tpl.Elements, &disk.Vector)
	}
	if err := flattenVMDisk(d, tpl); err != nil {
		t.Fatal(err)
	}
	disks := d.Get("disk").([]interface{})
	if len(disks) != 2 || disks[0].(map[string]interface{})["tf_uuid"] != "first" || disks[1].(map[string]interface{})["tf_uuid"] != "second" {
		t.Fatalf("external disk adopted or configured order lost: %#v", disks)
	}
	duplicate := testVMDisk(3, "first", 400, "vdd")
	tpl.Elements = append(tpl.Elements, &duplicate.Vector)
	if err := flattenVMDisk(d, tpl); err == nil {
		t.Fatal("duplicate remote UUID should report an error")
	}
}

func TestVMDiskUUIDPlan(t *testing.T) {
	r := resourceOpennebulaVirtualMachine()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"name": "test", "cpu": 1.0, "memory": 128,
		"disk": []interface{}{map[string]interface{}{"image_id": 3, "size": 100, "tf_uuid": "original"}},
	})
	d.SetId("9")
	state := d.State()
	state.Attributes["vmgroup.#"] = "0"
	for _, tc := range []struct {
		name        string
		image, size int
	}{
		{"unchanged", 3, 100},
		{"resize", 3, 200},
		{"replace image", 4, 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := terraform.NewResourceConfigRaw(map[string]interface{}{
				"name": "test", "cpu": 1.0, "memory": 128,
				"disk": []interface{}{map[string]interface{}{"image_id": tc.image, "size": tc.size}},
			})
			diff, err := r.SimpleDiff(context.Background(), state, config, &Configuration{})
			if err != nil {
				t.Fatal(err)
			}
			attr := diff.Attributes["disk.0.tf_uuid"]
			if attr != nil {
				t.Fatalf("existing UUID should be preserved: %+v", attr)
			}
		})
	}
	t.Run("add disk", func(t *testing.T) {
		config := terraform.NewResourceConfigRaw(map[string]interface{}{
			"name": "test", "cpu": 1.0, "memory": 128,
			"disk": []interface{}{
				map[string]interface{}{"image_id": 4, "size": 100},
				map[string]interface{}{"image_id": 3, "size": 100},
			},
		})
		diff, err := r.SimpleDiff(context.Background(), state, config, &Configuration{})
		if err != nil {
			t.Fatal(err)
		}
		if attr := diff.Attributes["disk.0.tf_uuid"]; attr != nil {
			t.Fatalf("existing UUID should remain stable when adding a disk: %+v", attr)
		}
	})
}

func TestVMDiskUUIDChanges(t *testing.T) {
	old := []interface{}{map[string]interface{}{"image_id": 3, "tf_uuid": "original"}}
	new := []interface{}{
		map[string]interface{}{"image_id": 4, "tf_uuid": "original"},
		map[string]interface{}{"image_id": 3},
	}
	detach, attach, _ := diffVMDiskConfig(new, old)
	if len(detach) != 1 || len(attach) != 2 {
		t.Fatalf("inserting a disk must not duplicate an existing UUID: detach=%v attach=%v", detach, attach)
	}
}

func TestVMDiskDiffPreservesDuplicates(t *testing.T) {
	first := map[string]interface{}{"image_id": 3, "size": 16, "disk_id": 1}
	second := map[string]interface{}{"image_id": 3, "size": 16, "disk_id": 2}
	new := []interface{}{
		map[string]interface{}{"image_id": 3, "size": 16},
		map[string]interface{}{"image_id": 3, "size": 16},
	}
	detach, attach, resize := diffVMDiskConfig(new, nil)
	if len(detach) != 0 || len(resize) != 0 || len(attach) != 2 {
		t.Fatalf("both identical disks must be attached: detach=%v attach=%v resize=%v", detach, attach, resize)
	}
	new[1].(map[string]interface{})["size"] = 32
	detach, attach, resize = diffVMDiskConfig(new, []interface{}{first, second})
	if len(detach) != 0 || len(attach) != 0 || len(resize) != 1 || resize[0].attached["disk_id"] != 2 {
		t.Fatalf("only the second disk should resize: detach=%v attach=%v resize=%v", detach, attach, resize)
	}
	detach, attach, resize = diffVMDiskConfig(new[:1], []interface{}{first, second})
	if len(detach) != 1 || detach[0]["disk_id"] != 2 || len(attach) != 0 || len(resize) != 0 {
		t.Fatalf("only the second disk should detach: detach=%v attach=%v resize=%v", detach, attach, resize)
	}
}
