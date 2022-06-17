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
	// States management

	// Creation
	vmCreateTransientStates = VMStates{
		States: []vm.State{vm.Pending},
		LCMs:   []vm.LCMState{vm.Prolog, vm.Boot},
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

	vmResizeTransientStates = VMStates{
		States: []vm.State{vm.Active},
		LCMs:   []vm.LCMState{vm.HotplugResize},
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

func NewVMStateConf(timeout time.Duration, target, pending []string) resource.StateChangeConf {
	return resource.StateChangeConf{
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
}

func NewVMUpdateStateConf(timeout time.Duration, target, pending []string) resource.StateChangeConf {
	return resource.StateChangeConf{
		Delay:      5 * time.Second,
		MinTimeout: 2 * time.Second,
	}
}

type VMStates struct {
	States []vm.State
	LCMs   []vm.LCMState
}

func NewVMState(states ...vm.State) VMStates {
	vmStates := VMStates{}
	copy(vmStates.States, states)
	return vmStates
}

func NewVMLCMState(lcms ...vm.LCMState) VMStates {
	vmStates := VMStates{}
	copy(vmStates.LCMs, lcms)
	return vmStates
}

func (s VMStates) ToStrings() []string {
	ret := make([]string, len(s.States)+len(s.LCMs), 0)
	for _, state := range s.States {
		ret = append(ret, fmt.Sprintf("%s:%s", state, vm.LcmInit))
	}
	for _, lcm := range s.States {
		ret = append(ret, fmt.Sprintf("%s:%s", vm.Active, lcm))
	}
	return ret
}

func (s VMStates) Append(states VMStates) VMStates {
	s.States = append(s.States, states.States...)
	s.LCMs = append(s.LCMs, states.LCMs...)
	return s
}

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

		vmState, vmLcmState, err := vmInfos.State()
		if err != nil {
			return vmInfos, "err_unknown_state", err
		}
		log.Printf("VM (ID:%d, name:%s) is currently in state %s and in LCM state %s", vmInfos.ID, vmInfos.Name, vmState.String(), vmLcmState.String())

		vmFullState := fmt.Sprintf("%s:%s", vmState, vmLcmState)

		// In case we are in some failure state, we try to retrieve more error informations from the vm user template
		if vmState == vm.Active &&
			vmLcmState == vm.BootFailure ||
			vmLcmState == vm.PrologFailure ||
			vmLcmState == vm.EpilogFailure {
			vmerr, _ := vmInfos.UserTemplate.Get(vmk.Error)
			return vmInfos, vmFullState, fmt.Errorf("VM (ID:%d) entered fail state, error: %s", vmInfos.ID, vmerr)
		}

		return vmInfos, vmFullState, nil
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
