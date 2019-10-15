package metrics

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/opennebula-go/api"
	vm "github.com/megamsys/opennebula-go/virtualmachine"
	"strconv"
	"time"
)

const (
	DISKS  = "disks"
	MEMORY = "memory"
	CPU    = "cpu"
)

type Accounting struct {
	Api       *api.Rpc
	StartTime int64
	EndTime   int64
}

type History struct {
	HostName string `xml:"HOSTNAME"`
	Stime    int64  `xml:"STIME"`
	Etime    int64  `xml:"ETIME"`
	PStime   int64  `xml:"PSTIME"`
	PEtime   int64  `xml:"PETIME"`
	RStime   int64  `xml:"RSTIME"`
	REtime   int64  `xml:"RETIME"`
	VM       *VM    `xml:"VM"`
}

type VM struct {
	Name      string    `xml:"NAME"`
	State     string    `xml:"STATE"`
	Lcm_state string    `xml:"LCM_STATE"`
	Stime     int64     `xml:"STIME"`
	Etime     int64     `xml:"ETIME"`
	Template  *Template `xml:"TEMPLATE"`
}

type Template struct {
	Context     Context `xml:"CONTEXT"`
	Cpu         string  `xml:"CPU"`
	Cpu_cost    string  `xml:"CPU_COST"`
	Disks       []Disk  `xml:"DISK"`
	Vcpu        string  `xml:"VCPU"`
	Memory      string  `xml:"MEMORY"`
	Memory_cost string  `xml:"MEMORY_COST"`
	Disk_cost   string  `xml:"DISK_COST"`
}

type Disk struct {
	DiskId string `xml:"DISK_ID"`
	Size   int64  `xml:"SIZE"`
}

type Context struct {
	Name          string `xml:"NAME"`
	Accounts_id   string `xml:"ACCOUNTS_ID"`
	Assembly_id   string `xml:"ASSEMBLY_ID"`
	Assemblies_id string `xml:"ASSEMBLIES_ID"`
	Quota_id      string `xml:"QUOTA_ID"`
}

type OpenNebulaStatus struct {
	History_Records []*History `xml:"HISTORY"`
}

func (a *Accounting) Get() (interface{}, error) {
	log.Debugf("showback Get (%d, %d) started", a.StartTime, a.EndTime)
	args := []interface{}{a.Api.Key, -2, a.StartTime, a.EndTime}
	res, err := a.Api.Call(api.VMPOOL_ACCOUNTING, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *History) Cpu() string {
	return h.VM.Template.Cpu
}

func (h *History) VCpu() string {
	return h.VM.Template.Vcpu
}

func (h *History) CpuCost() string {
	return h.VM.Template.Cpu_cost
}

func (h *History) Memory() string {
	return h.VM.Template.Memory
}

func (h *History) MemoryCost() string {
	return h.VM.Template.Memory_cost
}

func (h *History) DiskSize() int64 {
	var totalsize int64
	for _, v := range h.VM.Template.Disks {
		totalsize = totalsize + v.Size
	}
	return totalsize
}

func (h *History) Disks() []Disk {
	return h.VM.Template.Disks
}

func (h *History) DiskCost() string {
	return h.VM.Template.Disk_cost
}

func (h *History) AssemblyName() string {
	return h.VM.Name
}

func (h *History) AccountsId() string {
	return h.VM.Template.Context.Accounts_id
}

func (h *History) AssembliesId() string {
	return h.VM.Template.Context.Assemblies_id
}

func (h *History) QuotaId() string {
	return h.VM.Template.Context.Quota_id
}

func (h *History) AssemblyId() string {
	return h.VM.Template.Context.Assembly_id
}

func (h *History) State() string {
	return h.VM.stateString()
}

func (h *History) LcmState() string {
	return h.VM.lcmStateString()
}

func (h *History) Elapsed() string {
	return strconv.FormatFloat(time.Since(time.Unix(h.VM.Stime, 0)).Hours(), 'E', -1, 64)
}

func (v *VM) stateAsInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 22
}

func (v *VM) stateString() string {
	return vm.VmStateString[vm.VmState(v.stateAsInt(v.State))]
}

func (v *VM) lcmStateString() string {
	return vm.LcmStateString[vm.LcmState(v.stateAsInt(v.Lcm_state))]
}
