package clusters

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/megamsys/opennebula-go/api"
)

var ErrNoCL = errors.New("no cluster found, Did you create them ?")

const (
	GETCLUSTERS     = "one.clusterpool.info"
	GETCLUSTER      = "one.cluster.info"
	CREATE_CLUSTER  = "one.cluster.allocate"
	UPDATE_CLUSTER  = "one.cluster.update"
	CLUSTER_ADDHOST = "one.cluster.addhost"
	CLUSTER_ADDVNET = "one.cluster.addvnet"
	CLUSTER_ADD_DS  = "one.cluster.adddatastore"
	CLUSTER_DELHOST = "one.cluster.delhost"
	CLUSTER_DELVNET = "one.cluster.delvnet"
	CLUSTER_DEL_DS  = "one.cluster.deldatastore"
)

type Clusters struct {
	Clusters []*Cluster `xml:"CLUSTER"`
	T        *api.Rpc
}

type Cluster struct {
	Id         int        `xml:"ID"`
	Name       string     `xml:"NAME"`
	Hosts      *Host      `xml:"HOSTS"`
	Datastores *Datastore `xml:"DATASTORES"`
	Vnets      *Vnet      `xml:"VNETS"`
}

type Host struct {
	ID []*string `xml:"ID"`
}

type Datastore struct {
	ID []*string `xml:"ID"`
}

type Vnet struct {
	ID []*string `xml:"ID"`
}

func (c *Clusters) ClusterPoolinfo() (interface{}, error) {
	args := []interface{}{c.T.Key}
	res, err := c.T.Call(GETCLUSTERS, args)
	//close connection
	defer c.T.Client.Close()
	if err != nil {
		return nil, err
	}

	return res, nil

}

func (c *Clusters) ClusterInfo(cname string) (interface{}, error) {
	id, err := c.GetByName(cname)
	if err != nil {
		return nil, err
	}

	args := []interface{}{c.T.Key, id}
	res, err := c.T.Call(GETCLUSTER, args)
	//close connection
	defer c.T.Client.Close()
	if err != nil {
		return nil, err
	}

	return res, nil

}

func (c *Clusters) CreateCluster(name string) (interface{}, error) {
	args := []interface{}{c.T.Key, name}
	res, err := c.T.Call(CREATE_CLUSTER, args)
	//close connection
	defer c.T.Client.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Clusters) ClusterAddResources(method string, cid, rid int) (interface{}, error) {

	args := []interface{}{c.T.Key, cid, rid}
	res, err := c.T.Call(method, args)
	//close connection
	defer c.T.Client.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Given a name, this function will return the VM
func (c *Clusters) GetByName(name string) (int, error) {
	args := []interface{}{c.T.Key}

	res, err := c.T.Call(GETCLUSTERS, args)
	//close connection
	defer c.T.Client.Close()
	if err != nil {
		return -1, err
	}

	xmlCLS := &Clusters{}
	if err = xml.Unmarshal([]byte(res), xmlCLS); err != nil {
		fmt.Println(err)
	}

	for _, u := range xmlCLS.Clusters {
		if u.Name == name {
			return u.Id, nil
		}
	}

	return -1, ErrNoCL

}

func (c *Clusters) AddVnet(cls_id, vnet int) (interface{}, error) {
	args := []interface{}{c.T.Key, cls_id, vnet}
	res, err := c.T.Call(CLUSTER_ADDVNET, args)
	//close connection
	defer c.T.Client.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}
