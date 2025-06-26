package opennebula

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/template"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// generic dynamic template customization
type customDynTemplateFunc func(d *schema.ResourceData, tpl *dyn.Template) diag.Diagnostics

// resource customization
type customTemplateFunc func(ctx context.Context, d *schema.ResourceData, tpl *template.Template) diag.Diagnostics
type customVMFunc func(ctx context.Context, d *schema.ResourceData, tpl *vm.VM) diag.Diagnostics

type customFunc func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics

// vmDiskAttach is an helper that synchronously attach a disk
func vmDiskAttach(ctx context.Context, vmc *goca.VMController, timeout time.Duration, diskTpl *shared.Disk) (int, error) {

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

	// wait before checking disk list
	// final states ar added to transient one in case of slow cloud
	transient := vmDiskTransientStates
	transient.Append(vmDiskUpdateReadyStates)
	finalStrs := vmDiskUpdateReadyStates.ToStrings()
	stateConf := NewVMUpdateStateConf(timeout, transient.ToStrings(), finalStrs)

	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		return -1, fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(finalStrs, ","), err)
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
func vmDiskDetach(ctx context.Context, vmc *goca.VMController, timeout time.Duration, diskID int) error {

	log.Printf("[DEBUG] Detach disk %d", diskID)

	err := vmc.Disk(diskID).Detach()
	if err != nil {
		return fmt.Errorf("can't detach disk %d: %s\n", diskID, err)
	}

	// wait before checking disk list
	// final states ar added to transient one in case of slow cloud
	transient := vmDiskTransientStates.
		Append(vmDiskUpdateReadyStates)
	finalStrs := vmDiskUpdateReadyStates.ToStrings()
	stateConf := NewVMUpdateStateConf(timeout, transient.ToStrings(), finalStrs)

	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		return fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(finalStrs, ","), err)
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
func vmDiskResize(ctx context.Context, vmc *goca.VMController, timeout time.Duration, diskID, newsize int) error {

	log.Printf("[DEBUG] Resize disk %d", diskID)

	vmdc := vmc.Disk(diskID)

	err := vmdc.Resize(fmt.Sprintf("%d", newsize))
	if err != nil {
		return fmt.Errorf("can't resize image with Disk ID:%d: %s\n", diskID, err)
	}

	// wait before checking disk list
	// final states ar added to transient one in case of slow cloud
	transient := vmDiskTransientStates
	transient.Append(vmDiskResizeReadyStates)
	finalStrs := vmDiskResizeReadyStates.ToStrings()
	stateConf := NewVMUpdateStateConf(timeout, transient.ToStrings(), finalStrs)

	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		return fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(finalStrs, ","), err)
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
func vmNICAttach(ctx context.Context, vmc *goca.VMController, timeout time.Duration, nicTpl *shared.NIC) (int, error) {

	isNetworkMode := false
	networkMode, netModeErr := nicTpl.Get(shared.NetworkMode)
	networkID, netIDErr := nicTpl.GetI(shared.NetworkID)
	if netIDErr == nil {
		log.Printf("[DEBUG] Attach NIC to network (ID:%d)", networkID)
	} else {
		if netModeErr != nil {
			return -1, fmt.Errorf("NIC template neither have a network ID or a network mode")
		}
		log.Printf("[DEBUG] Attach NIC with network mode %s", networkMode)
		isNetworkMode = true
	}

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
		if isNetworkMode {
			return -1, fmt.Errorf("can't attach NIC (mode:%s): %s\n", networkMode, err)
		} else {
			return -1, fmt.Errorf("can't attach network (ID:%d): %s\n", networkID, err)
		}
	}

	// wait before checking NIC list
	// final states ar added to transient one in case of slow cloud
	transient := vmNICTransientStates
	transient.Append(vmNICUpdateReadyStates)
	finalStrs := vmNICUpdateReadyStates.ToStrings()
	stateConf := NewVMUpdateStateConf(timeout, transient.ToStrings(), finalStrs)

	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		return -1, fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(finalStrs, ","), err)
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

		if isNetworkMode {
			return -1, fmt.Errorf("network (mode:%s): %s\n", networkMode, vmerr)
		} else {
			return -1, fmt.Errorf("network (ID:%d): %s", networkID, vmerr)
		}

	} else {

		// If at least one nic has been updated, try to identify the one we just attached

	loop:
		for i, nic := range updatedNICs {

			for _, pair := range nicTpl.Pairs {

				value, err := nic.GetStr(pair.Key())
				if pair.Key() == "SECURITY_GROUPS" {
					// look for security group ID in the list
					ids := strings.Split(value, ",")
					found := false
					for _, id := range ids {

						if id != pair.Value {
							continue
						}
						found = true
						break
					}
					if !found {
						continue loop
					}
				} else if err != nil || value != pair.Value {
					continue loop
				}

			}

			attachedNIC = &updatedNICs[i]
			break
		}
		if attachedNIC == nil {

			if isNetworkMode {
				return -1, fmt.Errorf("network (mode %s): can't find the NIC\n", networkMode)
			} else {
				return -1, fmt.Errorf("network (ID:%d): can't find the NIC", networkID)
			}
		}
	}

	nicID, _ := attachedNIC.GetI(shared.NICID)

	return nicID, nil
}

// vmNICDetach is an helper that synchronously detach a NIC
func vmNICDetach(ctx context.Context, vmc *goca.VMController, timeout time.Duration, nicID int) error {

	log.Printf("[DEBUG] Detach NIC %d", nicID)

	err := vmc.DetachNIC(nicID)
	if err != nil {
		return fmt.Errorf("can't detach NIC %d: %s\n", nicID, err)
	}

	// wait before checking NIC list
	// final states ar added to transient one in case of slow cloud
	transient := vmNICTransientStates
	transient.Append(vmNICUpdateReadyStates)
	finalStrs := vmNICUpdateReadyStates.ToStrings()
	stateConf := NewVMUpdateStateConf(timeout, transient.ToStrings(), finalStrs)

	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		return fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(finalStrs, ","), err)
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

func vmNICAliasAttach(ctx context.Context, vmc *goca.VMController, timeout time.Duration, nicAliasTpl *shared.NIC) (int, error) {

	networkID, netIDErr := nicAliasTpl.GetI(shared.NetworkID)
	network, networkErr := nicAliasTpl.Get(shared.Network)

	if netIDErr != nil && networkErr != nil {
		return -1, fmt.Errorf("NIC_ALIAS template must have a 'NETWORK_ID or 'NETWORK' field")
	}
	if networkID > 0 {
		log.Printf("[DEBUG] Attach NIC Alias to network with ID:%d)", networkID)
	} else {
		log.Printf("[DEBUG] Attach NIC Alias to network %s", network)
	}

	// Retrieve NIC Alias list
	vm, err := vmc.Info(false)
	if err != nil {
		return -1, err
	}

	refNICAliases := schema.NewSet(schema.HashString, []any{})
	for _, nicAlias := range vm.Template.GetNICAliases() {
		refNICAliases.Add(nicAlias.String())
	}

	err = vmc.AttachNIC(nicAliasTpl.String())
	if err != nil {
		return -1, fmt.Errorf("can't attach NIC Alias (%v):\n %s", nicAliasTpl, err)
	}

	// wait before checking NIC list
	// final states ar added to transient one in case of slow cloud
	transient := vmNICTransientStates
	transient.Append(vmNICUpdateReadyStates)
	finalStrs := vmNICUpdateReadyStates.ToStrings()
	stateConf := NewVMUpdateStateConf(timeout, transient.ToStrings(), finalStrs)

	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		return -1, fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(finalStrs, ","), err)
	}

	// compare NIC list to check that a new NIC is attached
	vm, err = vmc.Info(false)
	if err != nil {
		return -1, err
	}

	updatedNICAliases := make([]shared.NIC, 0, 1)
	for _, nic := range vm.Template.GetNICAliases() {

		if refNICAliases.Contains(nic.String()) {
			continue
		}

		updatedNICAliases = append(updatedNICAliases, nic)
	}

	var attachedNICAlias *shared.NIC

	if len(updatedNICAliases) == 0 {

		// If nic alias not attached, retrieve the VM error message
		vmerr, _ := vm.UserTemplate.Get(vmk.Error)
		return -1, fmt.Errorf("network (ID: %d / Name: %s): %s", networkID, network, vmerr)

	} else {
		// If at least one nic has been updated, try to identify the one we just attached
		for _, updatedNicAlias := range updatedNICAliases {
			if findNicInTemplate(*nicAliasTpl, updatedNicAlias) {
				attachedNICAlias = &updatedNicAlias
				break
			}
		}
		if attachedNICAlias == nil {
			return -1, fmt.Errorf("network (ID: %d / Name: %s): can't find the NIC Alias", networkID, network)
		}
	}

	nicID, _ := attachedNICAlias.GetI(shared.NICID)

	return nicID, nil
}

func vmNICAliasDetach(ctx context.Context, vmc *goca.VMController, timeout time.Duration, nicAliasNicId int) error {

	log.Printf("[DEBUG] Detach NIC Alias with NIC ID %d", nicAliasNicId)

	err := vmc.DetachNIC(nicAliasNicId)
	if err != nil {
		return fmt.Errorf("can't detach NIC Alias %d: %s\n", nicAliasNicId, err)
	}

	// wait before checking NIC list
	// final states ar added to transient one in case of slow cloud
	transient := vmNICTransientStates
	transient.Append(vmNICUpdateReadyStates)
	finalStrs := vmNICUpdateReadyStates.ToStrings()
	stateConf := NewVMUpdateStateConf(timeout, transient.ToStrings(), finalStrs)

	_, err = waitForVMStates(ctx, vmc, stateConf)
	if err != nil {
		return fmt.Errorf(
			"waiting for virtual machine (ID:%d) to be in state %s: %s", vmc.ID, strings.Join(finalStrs, ","), err)
	}

	// Check that NIC is detached
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	detached := true
	for _, attachedNICAlias := range vm.Template.GetNICAliases() {

		attachedNICAliasNicID, _ := attachedNICAlias.ID()
		if attachedNICAliasNicID == nicAliasNicId {
			detached = false
			break
		}

	}

	if !detached {
		// If NIC Alias still attached, retrieve error message
		vmerr, _ := vm.UserTemplate.Get(vmk.Error)

		return fmt.Errorf("NIC %d: %s", nicAliasNicId, vmerr)
	}

	return nil
}

func findNicInTemplate(nicTemplate shared.NIC, nic shared.NIC) bool {
	for _, templatePair := range nicTemplate.Pairs {
		value, err := nic.GetStr(templatePair.Key())
		if err != nil {
			return false
		}
		if templatePair.Key() == string(shared.SecurityGroups) && !securityGroupIdsMatches(templatePair, value) {
			return false
		}
		if value != templatePair.Value {
			return false
		}
	}
	return true
}

func securityGroupIdsMatches(templatePair dyn.Pair, nicSgsStr string) bool {
	if templatePair.Key() != string(shared.SecurityGroups) {
		return false
	}

	templateSGIds := strings.Split(templatePair.Value, ",")
	nicSGIds := strings.Split(nicSgsStr, ",")

	return ArraysAreEqual(templateSGIds, nicSGIds)
}
