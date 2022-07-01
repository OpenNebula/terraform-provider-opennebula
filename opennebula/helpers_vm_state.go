package opennebula

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	// VM states

	// Ready states: the possible states the VM should be in to be able to trigger the action
	// Transient states: when the action has been triggered, the VM will be in a set of transient states.
	// Sometimes these collections of state are used:
	// - before triggering the action to ensure it won't fail
	// - after triggering the action to control that the VM will go back in an expected state

	// Creation
	vmCreateTransientStates = VMStates{
		States: []vm.State{vm.Pending},
		LCMs:   []vm.LCMState{vm.LcmInit, vm.Prolog, vm.Boot},
	}

	// Deletion: terminate the VM
	vmDeleteReadyStates = VMStates{
		States: []vm.State{vm.Hold, vm.Poweroff, vm.Stopped, vm.Undeployed, vm.Suspended, vm.Done},
		LCMs:   []vm.LCMState{vm.Running},
	}

	vmDeleteTransientStates = VMStates{
		LCMs: []vm.LCMState{vm.Shutdown, vm.Epilog},
	}

	// Update: when the VM is powered off
	vmPowerOffTransientStates = VMStates{
		LCMs: []vm.LCMState{vm.ShutdownPoweroff},
	}

	// Update: VM resize
	vmResizeTransientStates = VMStates{
		LCMs: []vm.LCMState{vm.HotplugResize},
	}

	vmResizeReadyStates = VMStates{
		States: []vm.State{vm.Poweroff, vm.Undeployed},
	}

	// Disk and NIC updates
	vmDiskUpdateReadyStates = VMStates{
		States: []vm.State{vm.Poweroff},
		LCMs:   []vm.LCMState{vm.Running},
	}

	vmNICUpdateReadyStates = vmDiskUpdateReadyStates

	vmDiskResizeReadyStates = VMStates{
		States: []vm.State{vm.Poweroff, vm.Undeployed},
		LCMs:   []vm.LCMState{vm.Running},
	}

	vmDiskTransientStates = VMStates{
		LCMs: []vm.LCMState{vm.Hotplug, vm.HotplugPrologPoweroff, vm.HotplugEpilogPoweroff, vm.DiskResize, vm.DiskResizePoweroff, vm.DiskResizeUndeployed},
	}

	vmNICTransientStates = VMStates{
		LCMs: []vm.LCMState{vm.HotplugNic, vm.HotplugNicPoweroff},
	}
)

// VMStates represents a collection of VM states
type VMStates struct {
	States []vm.State
	LCMs   []vm.LCMState
}

func NewVMState(states ...vm.State) VMStates {
	vmStates := VMStates{
		States: make([]vm.State, len(states)),
	}
	copy(vmStates.States, states)
	return vmStates
}

func NewVMLCMState(lcms ...vm.LCMState) VMStates {
	vmStates := VMStates{
		LCMs: make([]vm.LCMState, len(lcms)),
	}
	copy(vmStates.LCMs, lcms)
	return vmStates
}

func (s VMStates) Append(states VMStates) VMStates {
	s.States = append(s.States, states.States...)
	s.LCMs = append(s.LCMs, states.LCMs...)
	return s
}

func (s VMStates) ToStrings() []string {
	ret := make([]string, 0, len(s.States)+len(s.LCMs))
	for _, state := range s.States {
		ret = append(ret, state.String())
	}
	for _, lcm := range s.LCMs {
		ret = append(ret, lcm.String())
	}
	return ret
}

// NewVMStateConf initialize a state change struct
func NewVMStateConf(timeout time.Duration, pending, target []string) resource.StateChangeConf {
	return resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
}

// NewVMUpdateStateConf initialize a state change struct
func NewVMUpdateStateConf(timeout time.Duration, pending, target []string) resource.StateChangeConf {
	return resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 2 * time.Second,
	}
}

// waitForVMStates wait for a VM to reach some states
func waitForVMStates(ctx context.Context, vmc *goca.VMController, stateChangeConf resource.StateChangeConf) (interface{}, error) {

	stateChangeConf.Refresh = func() (interface{}, string, error) {

		log.Println("Refreshing VM state...")

		vmInfos, err := vmc.Info(false)
		if err != nil {
			if NoExists(err) {
				// Do not return an error here as it is excpected if the VM is already in DONE state
				// after its destruction
				return nil, "err_not_found", nil
			}
			return nil, "err", err
		}

		vmState, vmLCMState, err := vmInfos.State()
		if err != nil {
			return vmInfos, "err_unknown_state", err
		}
		log.Printf("VM (ID:%d, name:%s) is currently in state %s and in LCM state %s", vmInfos.ID, vmInfos.Name, vmState.String(), vmLCMState.String())

		if vmState == vm.Active {

			// In case we are in some failure state, we try to retrieve more error informations from the vm user template
			if vmLCMState == vm.BootFailure ||
				vmLCMState == vm.PrologFailure ||
				vmLCMState == vm.EpilogFailure {
				vmerr, _ := vmInfos.UserTemplate.Get(vmk.Error)
				return vmInfos, vmLCMState.String(), fmt.Errorf("VM (ID:%d) entered fail state, error: %s", vmInfos.ID, vmerr)
			}

			return vmInfos, vmLCMState.String(), nil

		}

		return vmInfos, vmState.String(), nil
	}

	return stateChangeConf.WaitForStateContext(ctx)

}

func waitForVMsStates(ctx context.Context, c *goca.Controller, vmIDs []int, stateChangeConf resource.StateChangeConf) ([]interface{}, []error) {

	errors := make([]error, 0)
	vmsInfos := make([]interface{}, 0)

	for _, id := range vmIDs {
		vmInfo, err := waitForVMStates(ctx, c.VM(id), stateChangeConf)
		if vmInfo != nil {
			vmsInfos = append(vmsInfos, vmInfo)
		}
		if err != nil {
			errors = append(errors, err)
		}
	}

	return vmsInfos, errors
}
