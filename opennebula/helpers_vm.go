package opennebula

import (
	"fmt"
	"log"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// vmDiskAttach is an helper that synchronously attach a disk
func vmDiskAttach(vmc *goca.VMController, timeout int, diskTpl *shared.Disk) (int, error) {

	log.Printf("[DEBUG] Attach disk to virtual machine (ID:%d)", vmc.ID)

	// Retrieve disk list
	vm, err := vmc.Info(false)
	if err != nil {
		return -1, err
	}

	set := schema.NewSet(schema.HashString, []interface{}{})
	for _, disk := range vm.Template.GetDisks() {
		set.Add(disk.String())
	}

	err = vmc.DiskAttach(diskTpl.String())
	if err != nil {
		return -1, fmt.Errorf("can't attach image to virtual machine (ID:%d): %s\n", vmc.ID, err)

	}

	// wait before checking disk
	_, err = waitForVMState(vmc, timeout, vmDiskUpdateReadyStates...)
	if err != nil {
		return -1, fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(vmDiskUpdateReadyStates, " "), err)
	}

	// compare disk list to check that a new disk is attached
	vm, err = vmc.Info(false)
	if err != nil {
		return -1, err
	}

	oldDisks := make([]shared.Disk, 0, 1)
	for _, disk := range vm.Template.GetDisks() {

		if set.Contains(disk.String()) {
			continue
		}

		oldDisks = append(oldDisks, disk)
	}

	var attachedDisk *shared.Disk

	switch len(oldDisks) {
	case 0:

		// If disk not attached, retrieve error message
		vmerr, _ := vm.UserTemplate.Get(vmk.Error)

		return -1, fmt.Errorf("virtual machine (ID:%d): %s", vmc.ID, vmerr)

	case 1:
		attachedDisk = &oldDisks[0]
	default:
	loop:
		for i, disk := range oldDisks {

			for _, pair := range diskTpl.Pairs {

				value, err := disk.GetStr(pair.Key())
				if err != nil {

				}

				if value != pair.Value {
					continue loop
				}
			}

			attachedDisk = &oldDisks[i]
			break
		}
		if attachedDisk == nil {
			return -1, fmt.Errorf("can't find the disk attached to the virtual machine (ID:%d)", vmc.ID)
		}
	}

	diskID, _ := attachedDisk.GetI(shared.DiskID)

	return diskID, nil
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

// vmDiskResize is an helper that synchronously resize a disk
func vmDiskResize(vmc *goca.VMController, timeout, diskID, newsize int) error {

	log.Printf("[DEBUG] Resize disk %d", diskID)

	vmdc := vmc.Disk(diskID)

	err := vmdc.Resize(fmt.Sprintf("%d", newsize))
	if err != nil {
		return fmt.Errorf("can't resize image with Disk ID:%d: %s\n", diskID, err)
	}

	// wait before checking disk
	_, err = waitForVMState(vmc, timeout, vmDiskResizeReadyStates...)
	if err != nil {
		return fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(vmDiskUpdateReadyStates, " "), err)
	}

	// Check that disk has new size
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	for _, disks := range vm.Template.GetDisks() {

		vmDiskID, _ := disks.GetI(shared.DiskID)
		diskSize, _ := disks.GetI(shared.Size)
		if vmDiskID == diskID && diskSize == newsize {
			return nil
		}
	}

	// If error occured, retrieve error message
	vmerr, _ := vm.UserTemplate.Get(vmk.Error)

	return fmt.Errorf("image %d: %s", diskID, vmerr)
}

// vmNICAttach is an helper that synchronously attach a nic
func vmNICAttach(vmc *goca.VMController, timeout int, nicTpl *shared.NIC) (int, error) {

	networkID, err := nicTpl.GetI(shared.NetworkID)
	if err != nil {
		return -1, fmt.Errorf("NIC template doesn't have a network ID")
	}

	log.Printf("[DEBUG] Attach NIC to network (ID:%d)", networkID)

	// Retrieve NIC list
	vm, err := vmc.Info(false)
	if err != nil {
		return -1, err
	}

	refNICs := schema.NewSet(schema.HashString, []interface{}{})
	for _, nic := range vm.Template.GetNICs() {
		refNICs.Add(nic.String())
	}

	err = vmc.AttachNIC(nicTpl.String())
	if err != nil {
		return -1, fmt.Errorf("can't attach network with ID:%d: %s\n", networkID, err)
	}

	// wait before checking NIC
	_, err = waitForVMState(vmc, timeout, vmNICUpdateReadyStates...)
	if err != nil {
		return -1, fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(vmNICUpdateReadyStates, " "), err)
	}

	// compare NIC list to check that a new NIC is attached
	vm, err = vmc.Info(false)
	if err != nil {
		return -1, err
	}

	updatedNICs := make([]shared.NIC, 0, 1)
	for _, nic := range vm.Template.GetNICs() {

		if refNICs.Contains(nic.String()) {
			continue
		}

		updatedNICs = append(updatedNICs, nic)
	}

	var attachedNIC *shared.NIC

	if len(updatedNICs) == 0 {

		// If nic not attached, retrieve the VM error message

		vmerr, _ := vm.UserTemplate.Get(vmk.Error)

		return -1, fmt.Errorf("network %d: %s", networkID, vmerr)

	} else {

		// If at least one nic has been updated, try to identify the one we just attached

	loop:
		for i, nic := range updatedNICs {

			for _, pair := range nicTpl.Pairs {

				value, err := nic.GetStr(pair.Key())
				if err != nil {

				}

				if value != pair.Value {
					continue loop
				}
			}

			attachedNIC = &updatedNICs[i]
			break
		}
		if attachedNIC == nil {
			return -1, fmt.Errorf("network %d: can't find the nic", networkID)
		}
	}

	nicID, _ := attachedNIC.GetI(shared.NICID)

	return nicID, nil
}

// vmNICDetach is an helper that synchronously detach a NIC
func vmNICDetach(vmc *goca.VMController, timeout int, nicID int) error {

	log.Printf("[DEBUG] Detach NIC %d", nicID)

	err := vmc.DetachNIC(nicID)
	if err != nil {
		return fmt.Errorf("can't detach NIC %d: %s\n", nicID, err)
	}

	// wait before checking NIC
	_, err = waitForVMState(vmc, timeout, vmNICUpdateReadyStates...)
	if err != nil {
		return fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(vmNICUpdateReadyStates, " "), err)
	}

	// Check that NIC is detached
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	detached := true
	for _, attachedNIC := range vm.Template.GetNICs() {

		attachedNICID, _ := attachedNIC.ID()
		if attachedNICID == nicID {
			detached = false
			break
		}

	}

	if !detached {
		// If NIC still attached, retrieve error message
		vmerr, _ := vm.UserTemplate.Get(vmk.Error)

		return fmt.Errorf("NIC %d: %s", nicID, vmerr)
	}

	return nil
}
