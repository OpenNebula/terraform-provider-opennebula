package compute

import (
	"encoding/xml"
	"errors"
	"github.com/megamsys/opennebula-go/api"
	"github.com/megamsys/opennebula-go/template"
	"github.com/megamsys/opennebula-go/virtualmachine"
)

var (
	ErrNoVM = errors.New("no vm found, Did you launch them ?")
)

const (
	RECOVER_FORCE_DELETE = 3

	DELETE        = "terminate"
	REBOOT        = "reboot"
	POWEROFF      = "poweroff"
	RESUME        = "resume"
	SUSPEND       = "suspend"
	FORCE_DELETE  = "terminate-hard"
	UNDEPLOY_HARD = "undeploy-hard"
	UNDEPLOY      = "undeploy"
	POWEROFF_HARD = "poweroff-hard"
	REBOOT_HARD   = "reboot-hard"

	ASSEMBLY_ID    = "assembly_id"
	ASSEMBLIES_ID  = "assemblies_id"
	ACCOUNTS_ID    = "accounts_id"
	ORG_ID         = "org_id"
	API_KEY        = "api_key"
	QUOTA_ID       = "quota_id"
	PLATFORM_ID    = "platform_id"
	SSH_PUBLIC_KEY = "SSH_PUBLIC_KEY"
)

type VirtualMachine struct {
	Name         string
	TemplateName string
	TemplateId   int
	Image        string
	ContextMap   map[string]string
	Cpu          string
	CpuCost      string
	VCpu         string
	Memory       string
	MemoryCost   string
	HDD          string
	HDDCost      string
	Files        string
	Region       string
	ClusterId    string
	VMId         int
	Vnets        map[string]string
	ForceNetwork bool
	T            *api.Rpc
}

type Image struct {
	Name    string
	VMId    int
	DiskId  int
	SnapId  int
	ImageId int
	Region  string
	T       *api.Rpc
}

// Creates a new VirtualMachine
func (v *VirtualMachine) Compute() (template.UserTemplates, error) {
	finalXML := template.UserTemplates{}
	templateObj := template.TemplateReqs{TemplateName: v.TemplateName, T: v.T}
	XMLtemplate, err := templateObj.Get()
	if err != nil {
		return finalXML, err
	}

	XMLtemplate[0].Template.Cpu = v.Cpu
	XMLtemplate[0].Template.VCpu = v.VCpu
	XMLtemplate[0].Template.Memory = v.Memory

	XMLtemplate[0].Template.Cpu_cost = v.CpuCost
	XMLtemplate[0].Template.Memory_cost = v.MemoryCost
	XMLtemplate[0].Template.Disk_cost = v.HDDCost
	XMLtemplate[0].Template.Context.Accounts_id = v.ContextMap[ACCOUNTS_ID]
	XMLtemplate[0].Template.Context.Platform_id = v.ContextMap[PLATFORM_ID]
	XMLtemplate[0].Template.Context.Assembly_id = v.ContextMap[ASSEMBLY_ID]
	XMLtemplate[0].Template.Context.Assemblies_id = v.ContextMap[ASSEMBLIES_ID]
	XMLtemplate[0].Template.Context.Quota_id = v.ContextMap[QUOTA_ID]
	XMLtemplate[0].Template.Context.ApiKey = v.ContextMap[API_KEY]
	XMLtemplate[0].Template.Context.Org_id = v.ContextMap[ORG_ID]
	XMLtemplate[0].Template.Context.SSH_Public_key = v.ContextMap[SSH_PUBLIC_KEY]

	if v.Files != "" {
		XMLtemplate[0].Template.Context.Files = v.Files
	}
	if len(XMLtemplate[0].Template.Disks) > 0 {
		if XMLtemplate[0].Template.Disks[0] != nil {
			if v.HDD != "" {
				XMLtemplate[0].Template.Disks[0].Size = v.HDD
			}
			if v.Image != "" {
				XMLtemplate[0].Template.Disks[0].Image = v.Image
			}
		}
	}

	if len(v.ClusterId) > 0 {
		XMLtemplate[0].Template.Sched_requirments = "CLUSTER_ID=\"" + v.ClusterId + "\""
	}

	if len(v.Vnets) > 0 {
		XMLtemplate[0].Template.Nic = XMLtemplate[0].Template.Nic[:0]
		for _, v := range v.Vnets {
			net := &template.NIC{Network: v, Network_uname: "oneadmin"}
			XMLtemplate[0].Template.Nic = append(XMLtemplate[0].Template.Nic, net)
		}
	}

	finalXML.UserTemplate = XMLtemplate
	return finalXML, nil
}

func (v *VirtualMachine) Create(tmp template.UserTemplates) (interface{}, error) {
	finalData, _ := xml.Marshal(tmp.UserTemplate[0].Template)
	data := string(finalData)
	args := []interface{}{v.T.Key, tmp.UserTemplate[0].Id, v.Name, false, data}

	return v.T.Call(api.TEMPLATE_INSTANTIATE, args)
}

/**
*
* Actions of virtualMachine
* boot ,terminate, suspend, hold, stop, resume, release, poweroff, reboot
*
**/

func (v *VirtualMachine) actions(action string) (interface{}, error) {
	args := []interface{}{v.T.Key, action, v.VMId}
	return v.T.Call(api.ONE_VM_ACTION, args)
}

/**
* REBoot a new virtualMachine
**/
func (v *VirtualMachine) Reboot() (interface{}, error) {
	return v.actions(REBOOT)
}

/**
* POWEROFF a new virtualMachine
**/
func (v *VirtualMachine) Poweroff() (interface{}, error) {
	return v.actions(POWEROFF)
}

/**
* Resume a new virtualMachine
**/
func (v *VirtualMachine) Resume() (interface{}, error) {
	return v.actions(RESUME)

}

/**
 * Deletes an existing virtualMachine
 **/
func (v *VirtualMachine) Delete() (interface{}, error) {
	return v.actions(DELETE)

}

/**
 * Suspends a virtualMachine
 **/
func (v *VirtualMachine) Suspends() (interface{}, error) {
	return v.actions(SUSPEND)
}

//  * Undeploy a virtualMachine
func (v *VirtualMachine) Undeploy() (interface{}, error) {
	return v.actions(UNDEPLOY)
}

//  * UndeployHard a virtualMachine
func (v *VirtualMachine) UndeployHard() (interface{}, error) {
	return v.actions(UNDEPLOY_HARD)
}

//  * PoweroffHard a virtualMachine

func (v *VirtualMachine) PoweroffHard() (interface{}, error) {
	return v.actions(POWEROFF_HARD)
}

//  * RebootHard a virtualMachine

func (v *VirtualMachine) RebootHard() (interface{}, error) {
	return v.actions(REBOOT_HARD)
}

//  * TerminateHard a virtualMachine

func (v *VirtualMachine) TerminateHard() (interface{}, error) {
	return v.actions(FORCE_DELETE)
}

/**
 * Deletes a new virtualMachine in ANY state (force delete)
 **/
func (v *VirtualMachine) RecoverDelete() (interface{}, error) {
	return v.T.Call(api.ONE_RECOVER, []interface{}{v.T.Key, v.VMId, RECOVER_FORCE_DELETE})
}

/**
 * VM save as a new Image (DISK_SNAPSHOT)
 **/

func (v *Image) DiskSaveAs() (interface{}, error) {
	args := []interface{}{v.T.Key, v.VMId, v.DiskId, v.Name, "", v.SnapId}
	return v.T.Call(api.ONE_DISK_SNAPSHOT, args)
}

func (v *Image) RemoveImage() (interface{}, error) {
	args := []interface{}{v.T.Key, v.ImageId}
	return v.T.Call(api.ONE_IMAGE_REMOVE, args)
}

func listByName(name string, client *api.Rpc) (*virtualmachine.UserVM, error) {
	vms := virtualmachine.Query{VMName: name, T: client}

	svm, err := vms.GetByName()
	if err != nil {
		return nil, err
	}

	if len(svm) <= 0 || svm[0] == nil {
		return nil, ErrNoVM
	}

	return svm[0], nil
}
