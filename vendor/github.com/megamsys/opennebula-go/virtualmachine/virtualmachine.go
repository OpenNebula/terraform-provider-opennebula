package virtualmachine

import (
	"encoding/xml"
	"github.com/megamsys/opennebula-go/api"
	"strconv"
	"strings"
)

const (
	//VmState starts at 0
	INIT VmState = iota
	PENDING
	HOLD
	ACTIVE
	STOPPED
	SUSPENDED
	DONE
	UNKNOWNSTATE
	POWEROFF
	UNDEPLOYED
)

type Query struct {
	VMName string
	VMId   int
	T      *api.Rpc
}

type UserVMs struct {
	UserVM []*UserVM `xml:"VM"`
}

type UserVM struct {
	Id   int    `xml:"ID"`
	Uid  int    `xml:"UID"`
	Name string `xml:"NAME"`
}

type Vnc struct {
	VmId string
	T    *api.Rpc
	VM   *VM `xml:"VM"`
}

type VM struct {
	Id             string          `xml:"ID"`
	Name           string          `xml:"NAME"`
	State          int             `xml:"STATE"`
	LcmState       int             `xml:"LCM_STATE"`
	VmTemplate     *VmTemplate     `xml:"TEMPLATE"`
	UserTemplate   UserTemplate    `xml:"USER_TEMPLATE"`
	HistoryRecords *HistoryRecords `xml:"HISTORY_RECORDS"`
	Snapshots      *Snapshots      `xml:"SNAPSHOTS"`
}

type VmTemplate struct {
	Graphics *Graphics `xml:"GRAPHICS"`
	Context  *Context  `xml:"CONTEXT"`
	Nics     []Nic     `xml:"NIC"`
}

type Nic struct {
	Network   string `xml:"NETWORK"`
	Id        string `xml:"NIC_ID"`
	IPaddress string `xml:"IP"`
	Mac       string `xml:"MAC"`
}

type Context struct {
	VMIP string `xml:"ETH0_IP"`
}

type HistoryRecords struct {
	History *History `xml:"HISTORY"`
}
type History struct {
	HostName string `xml:"HOSTNAME"`
}

type Graphics struct {
	Port string `xml:"PORT"`
}

type UserTemplate struct {
	Description        string `xml:"DESCRIPTION"`
	Error              string `xml:"ERROR"`
	Sched_Requirements string `xml:"SCHED_REQUIREMENTS"`
}

type Snapshots struct {
	DiskId   int        `xml:"DISK_ID"`
	Snapshot []Snapshot `xml:"SNAPSHOT"`
}

type Snapshot struct {
	Name string `xml:"NAME"`
	Id   int    `xml:"ID"`
	Size string `xml:"SIZE"`
}

func (v *Vnc) GetVm() (*VM, error) {
	intstr, _ := strconv.Atoi(v.VmId)
	args := []interface{}{v.T.Key, intstr}
	onevm, err := v.T.Call(api.VM_INFO, args)
	if err != nil {
		return nil, err
	}

	xmlVM := &VM{}
	if err = xml.Unmarshal([]byte(onevm), xmlVM); err != nil {
		return nil, err
	}
	return xmlVM, err
}

//have to release hold ips
func (v *Vnc) AttachNic(network, ip string) error {
	var forceIp string
	id, _ := strconv.Atoi(v.VmId)
	if len(ip) > 0 {
		forceIp = ", IP=\"" + ip + "\""
	}
	nic := "NIC = [ NETWORK=\"" + network + "\", NETWORK_UNAME=\"oneadmin\"" + forceIp + "]"
	args := []interface{}{v.T.Key, id, nic}
	_, err := v.T.Call(api.ONE_VM_ATTACHNIC, args)
	return err
}

func (v *Vnc) DetachNic(nic int) error {
	id, _ := strconv.Atoi(v.VmId)
	args := []interface{}{v.T.Key, id, nic}
	_, err := v.T.Call(api.ONE_VM_DETACHNIC, args)
	return err
}

func (u *VM) GetPort() string {
	return u.VmTemplate.Graphics.Port
}

func (u *VM) GetState() int {
	return u.State
}

func (u *VM) GetLcmState() int {
	return u.LcmState
}

func (u *VM) GetHostIp() string {
	return u.HistoryRecords.History.HostName
}

func (u *VM) GetVMIP() string {
	return u.VmTemplate.Context.VMIP
}

func (v *VM) StateString() string {
	return VmStateString[VmState(v.State)]
}

func (v *VM) Nics() []Nic {
	return v.VmTemplate.Nics
}

func (v *VM) LenSnapshots() int {
	if v.Snapshots != nil {
		return len(v.Snapshots.Snapshot)
	}
	return 0
}

func (v *VM) NetworkIdByIP(ip string) string {
	for _, n := range v.VmTemplate.Nics {
		if ip == n.IPaddress {
			return n.Id
		}
	}
	return ""
}

func (v *VM) LcmStateString() string {
	return LcmStateString[LcmState(v.LcmState)]
}

func (v *VM) IsFailure() bool {
	return strings.Contains(v.LcmStateString(), "failure")
}

func (v *VM) IsSnapshotReady() bool {
	return (v.State == int(ACTIVE) && v.LcmState == int(RUNNING)) || (v.State == int(POWEROFF) && v.LcmState == int(LCM_INIT))
}

// Given a name, this function will return the VM
func (v *Query) GetByName() ([]*UserVM, error) {
	args := []interface{}{v.T.Key, -2, -1, -1, -1}
	VMPool, err := v.T.Call(api.VMPOOL_INFO, args)
	if err != nil {
		return nil, err
	}

	xmlVM := UserVMs{}
	if err = xml.Unmarshal([]byte(VMPool), &xmlVM); err != nil {
		return nil, err
	}
	var matchedVM = make([]*UserVM, len(xmlVM.UserVM))

	for _, u := range xmlVM.UserVM {
		if u.Name == v.VMName {
			matchedVM[0] = u
		}
	}

	return matchedVM, nil

}
