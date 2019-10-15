package disk

import (
	"encoding/xml"
	"github.com/megamsys/opennebula-go/api"
)

type VmDisk struct {
	VmId int `xml:"ID"`
	Vm   Vm  `xml:"VM"`
	T    *api.Rpc
}

type Vm struct {
	VmTemplate VmTemplate `xml:"TEMPLATE"`
	Disk       Disk       `xml:"DISK"`
}
type VmTemplate struct {
	Disk []Disk `xml:"DISK"`
}
type Disk struct {
	Disk_Id    int    `xml:"DISK_ID"`
	Disk_Type  string `xml:"TYPE"`
	Image      string `xml:"IMAGE"`
	Dev_Prefix string `xml:"DEV_PREFIX"`
	Size       string `xml:"SIZE"`
	Target     string `xml:"TARGET"`
}

func (v *VmDisk) AttachDisk() (interface{}, error) {
	if v.Vm.Disk.Dev_Prefix == "" {
		v.Vm.Disk.Dev_Prefix = "vd"
	}
	if v.Vm.Disk.Disk_Type == "" {
		v.Vm.Disk.Disk_Type = "fs"
	}
	finalXML := VmDisk{}
	finalXML.Vm = v.Vm
	finalData, _ := xml.Marshal(finalXML.Vm)
	data := string(finalData)
	args := []interface{}{v.T.Key, v.VmId, data}
	res, err := v.T.Call(api.DISK_ATTACH, args)
	if err != nil {
		return nil, err
	}
	return res, err

}

func (v *VmDisk) DetachDisk() (interface{}, error) {
	args := []interface{}{v.T.Key, v.VmId, v.Vm.Disk.Disk_Id}
	res, err := v.T.Call(api.DISK_DETACH, args)
	if err != nil {
		return nil, err
	}
	return res, err

}

func (v *VmDisk) ListDisk() (*Vm, error) {
	args := []interface{}{v.T.Key, v.VmId}
	onevm, err := v.T.Call(api.VM_INFO, args)
	if err != nil {
		return nil, err
	}
	xmlVM := &Vm{}
	if err = xml.Unmarshal([]byte(onevm), xmlVM); err != nil {
		return nil, err
	}
	return xmlVM, err
}

func (u *Vm) GetDisks() []Disk {
	return u.VmTemplate.Disk
}

func (u *Vm) GetDiskIds() []int {
	var diskid []int
	for _, v := range u.VmTemplate.Disk {
		if v.Disk_Type == "fs" {
			diskid = append(diskid, v.Disk_Id)
		}
	}
	return diskid
}
