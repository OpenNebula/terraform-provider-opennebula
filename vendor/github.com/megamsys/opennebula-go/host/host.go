package host

import (
	// "encoding/xml"
	"github.com/megamsys/opennebula-go/api"
)

const (
	SHOW = "show"
)

type HQuery struct {
	Host *Host
	T    *api.Rpc
}

type Host struct {
	Id        int       `xml:"ID"`
	HostName  string    `xml:"NAME"`
	ClusterId int       `xml:"CLUSTER_ID"`
	IM_mad    string    `xml:"IM_MAD"`
	VMM_mad   string    `xml:"VMM_MAD"`
	Temp      *Template `xml:"TEMPLATE"`
}

type Template struct {
	VMs     []*VM  `xml:"VM"`
	Wildvms string `xml:"WILDS"`
}

type VM struct {
	Id        int    `xml:"ID"`
	Deploy_id int    `xml:"DEPLOY_ID"`
	VMName    string `xml:"VM_NAME"`
}

// Given a name, this function will return the VM
func (v *HQuery) HostInfos(a int) (interface{}, error) {
	args := []interface{}{v.T.Key, a}
	hostinfo, err := v.T.Call(api.ONE_HOST_INFO, args)
	if err != nil {
		return nil, err
	}

	return hostinfo, nil
}

func (v *HQuery) HostsPoolInfos(a int) (interface{}, error) {
	args := []interface{}{v.T.Key, a}
	hostInfos, err := v.T.Call(api.ONE_HOST_POOL, args)
	if err != nil {
		return nil, err
	}
	return hostInfos, nil
}

func (v *HQuery) AllocateHost() (interface{}, error) {
	args := []interface{}{v.T.Key, v.Host.HostName, v.Host.IM_mad, v.Host.VMM_mad, v.Host.ClusterId}
	addHost, err := v.T.Call(api.ONE_HOST_ALLOCATE, args)
	if err != nil {
		return nil, err
	}
	return addHost, nil
}

func (v *HQuery) DelHost(a int) (interface{}, error) {
	args := []interface{}{v.T.Key, a}
	delHost, err := v.T.Call(api.ONE_HOST_DELETE, args)
	if err != nil {
		return nil, err
	}
	return delHost, nil
}
