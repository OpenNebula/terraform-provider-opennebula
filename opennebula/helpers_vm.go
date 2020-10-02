package opennebula

import (
	"fmt"
	"log"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

// return disk configuration with image_id that only appear on refDisks side
func disksConfigDiff(refDisks, disks []interface{}) []map[string]interface{} {

	// get the list of disks ID to detach
	diffConfig := make([]map[string]interface{}, 0)

	for _, refDisk := range refDisks {
		refDiskConfig := refDisk.(map[string]interface{})
		refImageID := refDiskConfig["image_id"].(int)

		diff := true
		for _, disk := range disks {
			diskConfig := disk.(map[string]interface{})
			diskImageID := diskConfig["image_id"].(int)

			if refImageID == diskImageID {
				diff = false
				break
			}
		}

		if diff {
			diffConfig = append(diffConfig, refDiskConfig)
		}
	}

	return diffConfig
}

// vmDiskAttach is an helper that synchronously attach a disk
func vmDiskAttach(vmc *goca.VMController, timeout int, diskTpl *shared.Disk) error {

	imageID, err := diskTpl.GetI(shared.ImageID)
	if err != nil {
		return fmt.Errorf("disk template doesn't have and image ID")
	}

	log.Printf("[DEBUG] Attach image (ID:%d) as disk", imageID)

	err = vmc.DiskAttach(diskTpl.String())
	if err != nil {
		return fmt.Errorf("can't attach image with ID:%d: %s\n", imageID, err)
	}

	// wait before checking disk
	_, err = waitForVMState(vmc, timeout, vmDiskUpdateReadyStates...)
	if err != nil {
		return fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(vmDiskUpdateReadyStates, " "), err)
	}

	// Check that disk is attached
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	for _, attachedDisk := range vm.Template.GetDisks() {

		attachedDiskImageID, _ := attachedDisk.GetI(shared.ImageID)
		if attachedDiskImageID == imageID {
			return nil
		}
	}

	// If disk not attached, retrieve error message
	vmerr, _ := vm.UserTemplate.Get(vmk.Error)

	return fmt.Errorf("image %d: %s", imageID, vmerr)
}

// vmDiskDetach is an helper that synchronously detach a disk
func vmDiskDetach(vmc *goca.VMController, timeout int, diskID int) error {

	log.Printf("[DEBUG] Detach disk %d", diskID)

	err := vmc.Disk(diskID).Detach()
	if err != nil {
		return fmt.Errorf("can't detach disk %d: %s\n", diskID, err)
	}

	// wait before checking disk
	_, err = waitForVMState(vmc, timeout, vmDiskUpdateReadyStates...)
	if err != nil {
		return fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(vmDiskUpdateReadyStates, " "), err)
	}

	// Check that disk is detached
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	detached := true
	for _, attachedDisk := range vm.Template.GetDisks() {

		attachedDiskID, _ := attachedDisk.ID()
		if attachedDiskID == diskID {
			detached = false
			break
		}

	}

	if !detached {
		// If disk still attached, retrieve error message
		vmerr, _ := vm.UserTemplate.Get(vmk.Error)

		return fmt.Errorf("disk %d: %s", diskID, vmerr)
	}

	return nil
}
