package datastore

import (
	"encoding/xml"
	"github.com/megamsys/opennebula-go/api"
)

type DatastoreTemplate struct {
	Template Datastore `xml:"DATASTORES"`
	T        *api.Rpc
}

type Datastore struct {
	Id          int    `xml:"ID"`
	Name        string `xml:"NAME"`
	Ds_mad      string `xml:"DS_MAD"`
	Tm_mad      string `xml:"TM_MAD"`
	Disk_type   string `xml:"DISK_TYPE"`
	Bridge_list string `xml:"BRIDGE_LIST"`
	Ceph_host   string `xml:"CEPH_HOST"`
	Type        string `xml:"TYPE"`
	Safe_dirs   string `xml:"SAFE_DIRS"`
	Pool_name   string `xml:"Pool_NAME"`
	Ceph_user   string `xml:"CEPH_USER"`
	Ceph_secret string `xml:"CEPH_SECRET"`
	Host        string `xml:"HOST"`
	Vg_name     string `xml:"VG_NAME"`
}

func (v *DatastoreTemplate) AllocateDatastore(cluster_id int) (interface{}, error) {
	finalXML := DatastoreTemplate{}
	finalXML.Template = v.Template
	finalData, _ := xml.Marshal(finalXML.Template)
	data := string(finalData)
	args := []interface{}{v.T.Key, data, cluster_id}
	res, err := v.T.Call(api.ONE_DATASTORE_ALLOCATE, args)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (v *DatastoreTemplate) GetDATAs(a int) (interface{}, error) {
	args := []interface{}{v.T.Key, a}
	datastore, err := v.T.Call(api.ONE_DATASTORE_INFO, args)
	if err != nil {
		return nil, err
	}
	return datastore, nil

}

func (v *DatastoreTemplate) GetALL() (interface{}, error) {
	args := []interface{}{v.T.Key}
	datastores, err := v.T.Call(api.ONE_DATASTOREPOOL_INFO, args)
	if err != nil {
		return nil, err
	}

	return datastores, nil

}
