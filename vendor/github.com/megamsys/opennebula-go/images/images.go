package images

import (
	"encoding/xml"
	"fmt"
	"github.com/megamsys/opennebula-go/api"
)

type ImageType string

const (
	LOCKED  = 4
	READY   = 1
	USED    = 2
	FAILURE = 5
)

const (
	OPERATING_SYSTEM = ImageType("OS")
	CD_ROM           = ImageType("CDROM")
	DATABLOCK        = ImageType("DATABLOCK")
)

type Images struct {
	Images []Image `xml:"IMAGE"`
	T      *api.Rpc
}

type Image struct {
	Id          int       `xml:"ID"`
	Uid         int       `xml:"UID"`
	Gid         int       `xml:"GID"`
	Name        string    `xml:"NAME"`
	Uname       string    `xml:"UNAME"`
	Gname       string    `xml:"GNAME"`
	Type        ImageType `xml:"TYPE"`
	RegTime     string    `xml:"REG"`
	Size        int       `xml:"SIZE"`
	State       int       `xml:"STATE"`
	Source      string    `xml:"SOURCE"`
	Path        string    `xml:"PATH"`
	Persistent  string    `xml:"PERSISTENT"`
	DatastoreID int       `xml:"DATASTORE_ID"`
	Datastore   string    `xml:"DATASTORE"`
	FsType      string    `xml:"FSTYPE"`
	RunningVMs  int       `xml:"RUNNING_VMS"`
	VMs         Vms       `xml:"VMS"`
	T           *api.Rpc  `xml:"-"`
}

type Vms struct {
	Id []int `xml:"ID"`
}

func (v *Image) Create() (interface{}, error) {
	// qcow2 has some feature block so we use raw by default
	v.FsType = "raw"
	finalData, _ := xml.Marshal(v)
	data := string(finalData)
	args := []interface{}{v.T.Key, data, v.DatastoreID}
	return v.T.Call(api.ONE_IMAGE_CREATE, args)
}

func (v *Image) ByName() (*Image, error) {
	ims, err := v.List()
	if err != nil {
		return nil, err
	}
	for _, k := range ims.Images {
		if k.Name == v.Name {
			return &k, nil
		}
	}
	return nil, fmt.Errorf("ONE doesn't have any images name (%s)", v.Name)
}

func (v *Image) Delete() (interface{}, error) {
	args := []interface{}{v.T.Key, v.Id}
	return v.T.Call(api.ONE_IMAGE_DELETE, args)
}

func (v *Image) ChPersistent(state bool) (interface{}, error) {
	args := []interface{}{v.T.Key, v.Id, state}
	return v.T.Call(api.ONE_IMAGE_PERSISTENT, args)
}

func (v *Image) ChType() (interface{}, error) {
	args := []interface{}{v.T.Key, v.Id, string(v.Type)}
	return v.T.Call(api.ONE_IMAGE_TYPECHANGE, args)
}

func (v *Image) Rename(new_name string) (interface{}, error) {
	args := []interface{}{v.T.Key, v.Id, new_name}
	return v.T.Call(api.ONE_IMAGE_RENAME, args)
}

func (v *Image) Enable(state string) (interface{}, error) {
	args := []interface{}{v.T.Key, v.Id, state}
	return v.T.Call(api.ONE_IMAGE_ENABLE, args)
}

func (v *Image) Show() (*Image, error) {
	args := []interface{}{v.T.Key, v.Id}
	res, err := v.T.Call(api.ONE_IMAGE_SHOW, args)
	if err != nil {
		return nil, err
	}
	xmlImage := &Image{}
	if err = xml.Unmarshal([]byte(res), xmlImage); err != nil {
		return nil, err
	}
	return xmlImage, err
}

func (v *Image) List() (*Images, error) {
	first := -1 // -1 for default smaller ID
	last := -1  //-1 for default last ID
	args := []interface{}{v.T.Key, -1, first, last}
	res, err := v.T.Call(api.ONE_IMAGE_LIST, args)
	if err != nil {
		return nil, err
	}
	xmlImages := &Images{}
	if err = xml.Unmarshal([]byte(res), xmlImages); err != nil {
		return nil, err
	}
	return xmlImages, err
}

func (v *Image) State_string() string {
	switch v.State {
	case LOCKED:
		return "locked"
	case READY:
		return "ready"
	case USED:
		return "used"
	case FAILURE:
		return "failure"
	default:
		return "unknown"
	}
}
