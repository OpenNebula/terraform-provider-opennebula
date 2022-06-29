package opennebula

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	vn "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"
)

var vNetARAddInstancesStates = []string{"READY"}

func getARTemplate(AR *vn.AR) *vn.AddressRange {

	ARValue := reflect.ValueOf(*AR)
	typeOfAR := ARValue.Type()
	ARTemplate := vn.NewAddressRange()

	for i := 0; i < ARValue.NumField(); i++ {
		ARTemplate.AddPair(typeOfAR.Field(i).Name, ARValue.Field(i).Interface())
	}

	return ARTemplate
}

// vNetARAttach is an helper that synchronously attach a nic
func vNetARAdd(ctx context.Context, timeout time.Duration, vnc *goca.VirtualNetworkController, vNetID int, arTpl *vn.AddressRange) (int, error) {

	// Store reference AR list in a set
	vNetInfos, err := vnc.Info(false)
	if err != nil {
		return -1, err
	}

	refARs := schema.NewSet(schema.HashString, []interface{}{})
	for _, AR := range vNetInfos.ARs {
		tpl := getARTemplate(&AR)
		refARs.Add(tpl.String())
	}

	vNetInfos, err = vnc.Info(false)
	if err != nil {
		return -1, err
	}

	_, err = waitForVNetworkState(ctx, vnc, timeout, vNetARAddInstancesStates...)
	if err != nil {
		return -1, err
	}

	err = vnc.AddAR(arTpl.String())
	if err != nil {
		return -1, err
	}

	var attachedAR *vn.AddressRange

	err = resource.RetryContext(ctx, timeout, func() *resource.RetryError {

		vNetInfos, err := vnc.Info(false)
		if err != nil {
			return resource.RetryableError(err)
		}

		// list newly attached ARs
		updatedARs := make([]vn.AddressRange, 0, 1)
		for _, AR := range vNetInfos.ARs {

			tpl := getARTemplate(&AR)
			if refARs.Contains(tpl.String()) {
				continue
			}

			updatedARs = append(updatedARs, *tpl)
		}

		// check the retrieved list of ARs
		if len(updatedARs) == 0 {
			return resource.RetryableError(fmt.Errorf("virtual network (ID:%d): AR not attached", vNetID))
		} else {

			// If at least one nic has been updated, try to identify the one we just attached
		updatedARsLoop:
			for i, ar := range updatedARs {

				for _, pair := range arTpl.Pairs {

					value, err := ar.GetStr(pair.Key())
					if err != nil {
						continue updatedARsLoop
					}

					if value != pair.Value {
						continue updatedARsLoop
					}
				}

				attachedAR = &updatedARs[i]
				break
			}
			if attachedAR == nil {
				return resource.RetryableError(fmt.Errorf("virtual network (ID:%d): can't find the nic", vNetID))
			}

		}

		return nil
	})

	if err != nil {
		return -1, err
	}

	arID, _ := attachedAR.GetI("ID")

	return arID, nil
}

func isVRARAttached(vnc *goca.VirtualNetworkController, arID int) (bool, error) {

	vNetInfos, err := vnc.Info(false)
	if err != nil {
		return false, err
	}

	arIDStr := fmt.Sprint(arID)
	for _, attachedAR := range vNetInfos.ARs {

		if attachedAR.ID == arIDStr {
			return true, nil
		}

	}

	return false, nil
}

// vNetARDetach is an helper that synchronously detach a AR
func vNetARDetach(ctx context.Context, timeout time.Duration, controller *goca.Controller, vNetID int, arID int) error {

	vnc := controller.VirtualNetwork(vNetID)

	vNetInfos, err := vnc.Info(false)
	if err != nil {
		return err
	}

	// check if virtual network machines are in transient states
	ARs := vNetInfos.ARs
	if len(ARs) > 0 {
		_, err = waitForVNetworkState(ctx, vnc, timeout, vNetARAddInstancesStates...)
		if err != nil {
			return err
		}
	}

	err = vnc.FreeAR(arID)
	if err != nil {
		return fmt.Errorf("can't detach AR %d: %s\n", arID, err)
	}

	err = resource.RetryContext(ctx, timeout, func() *resource.RetryError {

		attached, err := isVRARAttached(vnc, arID)
		if err != nil {
			return resource.RetryableError(err)
		}

		if attached {
			return resource.RetryableError(fmt.Errorf("AR %d: not detached", arID))
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
