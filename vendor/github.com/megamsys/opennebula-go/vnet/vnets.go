package vnet

import (
	"encoding/xml"
	"fmt"
	"github.com/megamsys/opennebula-go/api"
	"strconv"
	"strings"
)

type VNETemplate struct {
	Template *Vnet `xml:"TEMPLATE"`
	T        *api.Rpc
}

type VNetPool struct {
	Vnets []*Vnet `xml:"VNET"`
	T     *api.Rpc
}

type Vnet struct {
	Id           int        `json:"id" xml:"ID"`
	Name         string     `json:"name" xml:"NAME"`
	Type         string     `json:"type" xml:"TYPE"`
	Description  string     `json:"description" xml:"DESCRIPTION"`
	Bridge       string     `json:"bridge" xml:"BRIDGE"`
	Network_addr string     `json:"network_addr" xml:"NETWORK_ADDRESS"`
	Network_mask string     `json:"network_mask" xml:"NETWORK_MASK"`
	Clusters     *Clusters  `json:"clusters" xml:"CLUSTERS"`
	Dns          string     `json:"dns" xml:"DNS"`
	Gateway      string     `json:"gateway" xml:"GATEWAY"`
	UsedIps      int        `json:"used_ips" xml:"USED_LEASES"`
	TotalIps     int        `json:"total_ips" xml:"TOTAL_IPS"`
	Ip_Leases    *Lease     `json:"ip_leases" xml:"LEASES"`
	Vn_mad       string     `json:"vn_mad" xml:"VN_MAD"`
	Addrs        []*Address `json:"addrs" xml:"AR"`
	AddrPool     *AddrPool  `json:"addr_pool" xml:"AR_POOL"`
}

type Lease struct {
	IP   string `json:"ip" xml:"IP"`
	Mac  string `json:"mac" xml:"MAC"`
	VmId string `json:"vm_id" xml:"VM"`
}
type Leases struct {
	Leases []Lease `json:"leases" xml:"LEASE"`
}
type Clusters struct {
	Id []string `json:"id" xml:"ID"`
}

type AddrPool struct {
	Addrs []*Address `json:"addrs" xml:"AR"`
}

type Address struct {
	Id      string    `json:"id" xml:"AR_ID"`
	Mac     string    `json:"mac" xml:"MAC"`
	Type    string    `json:"type" xml:"TYPE"`
	StartIP string    `json:"ip" xml:"IP"`
	Size    string    `json:"size" xml:"SIZE"`
	Leases  []*Leases `json:"leases" xml:"LEASES"`
}

func (v *VNETemplate) CreateVnet(cluster_id int) (interface{}, error) {
	finalXML := VNETemplate{}
	finalXML.Template = v.Template
	finalData, _ := xml.Marshal(finalXML.Template)
	data := string(finalData)
	args := []interface{}{v.T.Key, data, cluster_id}
	res, err := v.T.Call(api.VNET_CREATE, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *VNETemplate) VnetAddIps() (interface{}, error) {
	finalXML := VNETemplate{}
	finalXML.Template.Addrs = v.Template.Addrs
	finalData, _ := xml.Marshal(finalXML.Template)
	data := string(finalData)
	args := []interface{}{v.T.Key, data, v.Template.Id}
	res, err := v.T.Call(api.VNET_ADDIP, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *VNETemplate) VnetInfo(vnet_id int) (*Vnet, error) {
	args := []interface{}{v.T.Key, vnet_id}
	res, err := v.T.Call(api.VNET_SHOW, args)
	if err != nil {
		return nil, err
	}
	v.Template = &Vnet{}

	err = xml.Unmarshal([]byte(res), v.Template)
	return v.Template, nil
}

func (v *VNETemplate) VnetHold(vnet_id int, ip string) (interface{}, error) {
	return v.T.Call(api.VNET_HOLD, []interface{}{v.T.Key, vnet_id, "LEASES=[IP=" + ip + "]"})
}

func (v *VNETemplate) VnetRelease(vnet_id int, ip string) (interface{}, error) {
	return v.T.Call(api.VNET_RELEASE, []interface{}{v.T.Key, vnet_id, "LEASES=[IP=" + ip + "]"})
}

func (v *VNETemplate) VnetInfos(vnet_id []int) ([]*Vnet, error) {
	nets := make([]*Vnet, 0)
	for _, id := range vnet_id {
		net, err := v.VnetInfo(id)
		if err != nil {
			return nil, err
		}
		nets = append(nets, net)
	}
	return nets, nil
}

func (v *VNetPool) VnetPoolInfos(filter_id int) error {
	start_id := -1 //-1 for smaller values this is the offset used for pagination.
	end_id := -1   //-1 for get until the last ID
	args := []interface{}{v.T.Key, filter_id, start_id, end_id}
	res, err := v.T.Call(api.VNET_LIST, args)
	if err != nil {
		return err
	}
	err = xml.Unmarshal([]byte(res), v)
	v.setTotalIps()
	return err
}

func (v *VNetPool) FilletByType(t string) []*Vnet {
	vnets := make([]*Vnet, 0)
	for _, net := range v.Vnets {
		if t == net.AddrPool.Addrs[0].Type {
			vnets = append(vnets, net)
		}
	}
	return vnets
}

func (v *VNetPool) FilletByName(name string) (*Vnet, error) {
	for _, net := range v.Vnets {
		if net.Name == name {
			return net, nil
		}
	}
	return nil, fmt.Errorf("no such ( %s ) network available ", name)
}

func (v *VNetPool) setTotalIps() {
	for _, net := range v.Vnets {
		var total int
		for _, i := range net.AddrPool.Addrs {
			intstr, _ := strconv.Atoi(i.Size)
			total = total + intstr
		}
		net.TotalIps = total
	}
}

func (v *Vnet) IsUsed(ip string) bool {
	for _, addr := range v.AddrPool.Addrs {
		for _, leases := range addr.Leases {
			for _, lease := range leases.Leases {
				if lease.IP == strings.TrimSpace(ip) {
					return true
				}
			}
		}
	}
	return false
}
