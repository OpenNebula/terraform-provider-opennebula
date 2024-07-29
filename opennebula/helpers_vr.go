package opennebula

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
)

var vrNICDeleteInstancesStates = VMStates{
	States: []vm.State{vm.Poweroff, vm.Done},
	LCMs:   []vm.LCMState{vm.Running},
}

// vrNICAttach is an helper that synchronously attach a nic
func vrNICAttach(ctx context.Context, timeout time.Duration, controller *goca.Controller, vrID int, nicTpl *shared.NIC) (int, error) {

	networkID, err := nicTpl.GetI(shared.NetworkID)
	if err != nil {
		return -1, fmt.Errorf("NIC template doesn't have a network ID")
	}

	// Store reference NIC list in a set
	vrc := controller.VirtualRouter(vrID)
	vm, err := vrc.Info(false)
	if err != nil {
		return -1, err
	}

	refNICs := schema.NewSet(schema.HashString, []interface{}{})
	for _, nic := range vm.Template.GetNICs() {
		refNICs.Add(nic.String())
	}

	vrInfos, err := vrc.Info(false)
	if err != nil {
		return -1, err
	}

	// check if virtual router machines are in transient states
	if len(vrInfos.VMs.ID) > 0 {

		transient := vmNICTransientStates
		transient.Append(vmNICUpdateReadyStates)
		finalStrs := vmNICUpdateReadyStates.ToStrings()

		stateConf := NewVMUpdateStateConf(timeout,
			transient.ToStrings(),
			finalStrs,
		)

		_, errs := waitForVMsStates(ctx, controller, vrInfos.VMs.ID, stateConf)
		if len(errs) > 0 {
			var fullErr string
			for _, err := range errs {
				fullErr += fmt.Sprintf("\nVM error: %s\n", err.Error())
			}
			return -1, fmt.Errorf(
				"Virtual router waiting for virtual machines to be in state %s: %s", strings.Join(finalStrs, ","), fullErr)
		}
	}

	err = vrc.AttachNic(nicTpl.String())
	if err != nil {
		return -1, err
	}

	var attachedNIC *shared.NIC

	err = resource.RetryContext(ctx, timeout, func() *resource.RetryError {

		vrInfos, err := vrc.Info(false)
		if err != nil {
			return resource.RetryableError(err)
		}

		// list newly attached NICs
		updatedNICs := make([]shared.NIC, 0, 1)
		for _, nic := range vrInfos.Template.GetNICs() {

			if refNICs.Contains(nic.String()) {
				continue
			}

			updatedNICs = append(updatedNICs, nic)
		}

		// check the retrieved list of NICs
		if len(updatedNICs) == 0 {
			return resource.RetryableError(fmt.Errorf("virtual router (ID:%d): network %d: NIC not attached", vrID, networkID))
		} else {

			// If at least one nic has been updated, try to identify the one we just attached
		updatedNICsLoop:
			for i, nic := range updatedNICs {

				// For VRouter NICs, floating IPs are set using the "IP" field, but it is represented
				// as the "VROUTER_IP" field.
				if vrouterIP, err := nic.GetStr("VROUTER_IP"); err == nil {
					if _, err = nic.GetStr("IP"); err != nil {
						nic.Add("IP", vrouterIP)
					}
				}
				for _, pair := range nicTpl.Pairs {

					value, err := nic.GetStr(pair.Key())
					if err != nil {
						continue updatedNICsLoop
					}

					if value != pair.Value {
						continue updatedNICsLoop
					}
				}

				attachedNIC = &updatedNICs[i]
				break
			}
			if attachedNIC == nil {
				return resource.RetryableError(fmt.Errorf("virtual router (ID:%d): network %d: can't find the nic", vrID, networkID))
			}

		}

		return nil
	})

	if err != nil {
		return -1, err
	}

	nicID, _ := attachedNIC.GetI(shared.NICID)

	return nicID, nil

}

func isVRNICAttached(controller *goca.Controller, vrID, nicID int) (bool, error) {

	vrInfos, err := controller.VirtualRouter(vrID).Info(false)
	if err != nil {
		return false, err
	}

	for _, attachedNIC := range vrInfos.Template.GetNICs() {

		attachedNICID, _ := attachedNIC.ID()
		if attachedNICID == nicID {
			return true, nil
		}

	}

	return false, nil
}

// vrNICDetach is an helper that synchronously detach a NIC
func vrNICDetach(ctx context.Context, timeout time.Duration, controller *goca.Controller, vrID int, nicID int) error {

	vrc := controller.VirtualRouter(vrID)

	vrInfos, err := vrc.Info(false)
	if err != nil {
		return err
	}

	if len(vrInfos.VMs.ID) > 0 {

		// check if virtual router machines are in transient states
		// If one of the VMs is being deleted it will have it's NIC detached
		transient := vmNICTransientStates
		transient.Append(vmNICUpdateReadyStates)
		transient.Append(vmDeleteTransientStates)
		finalStrs := vrNICDeleteInstancesStates.ToStrings()

		stateConf := NewVMUpdateStateConf(timeout,
			transient.ToStrings(),
			finalStrs,
		)

		_, errs := waitForVMsStates(ctx, controller, vrInfos.VMs.ID, stateConf)
		if len(errs) > 0 {
			var fullErr string
			for _, err := range errs {
				fullErr += fmt.Sprintf("\nVM error: %s\n", err.Error())
			}
			return fmt.Errorf(
				"Virtual router waiting for virtual machines to be in state %s: %s", strings.Join(finalStrs, ","), fullErr)
		}
	}

	err = vrc.DetachNic(nicID)
	if err != nil {
		return fmt.Errorf("can't detach NIC %d: %s\n", nicID, err)
	}

	err = resource.RetryContext(ctx, timeout, func() *resource.RetryError {

		attached, err := isVRNICAttached(controller, vrID, nicID)
		if err != nil {
			return resource.RetryableError(err)
		}

		if attached {
			return resource.RetryableError(fmt.Errorf("NIC %d: not detached", nicID))
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
