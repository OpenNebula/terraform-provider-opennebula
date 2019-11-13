package api

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/kolo/xmlrpc"
	"github.com/megamsys/libgo/cmd"
	"strconv"
	"strings"
)

const (
	ENDPOINT        = "endpoint"
	USERID          = "userid"
	CURRENTUSER     = "username"
	PASSWORD        = "password"
	TEMPLATE        = "template"
	IMAGE           = "image"
	ONEZONE         = "region"
	VCPU_PERCENTAGE = "vcpu_percentage"
	CLUSTER         = "cluster"

	// user action methods
	ONE_USER_CREATE = "one.user.allocate"

	// virtual network action methods
	VNET_CREATE  = "one.vn.allocate"
	VNET_ADDIP   = "one.vn.add_ar"
	VNET_SHOW    = "one.vn.info"
	VNET_LIST    = "one.vnpool.info"
	VNET_HOLD    = "one.vn.hold"
	VNET_RELEASE = "one.vn.release"

	// host action methods
	ONE_HOST_INFO     = "one.host.info"
	ONE_HOST_POOL     = "one.hostpool.info"
	ONE_HOST_ALLOCATE = "one.host.allocate"
	ONE_HOST_DELETE   = "one.host.delete"

	// datastore action methods
	ONE_DATASTORE_INFO     = "one.datastore.info"
	ONE_DATASTOREPOOL_INFO = "one.datastorepool.info"
	ONE_DATASTORE_ALLOCATE = "one.datastore.allocate"

	// image action methods
	ONE_IMAGE_SHOW       = "one.image.info"
	ONE_IMAGE_LIST       = "one.imagepool.info"
	ONE_IMAGE_CREATE     = "one.image.allocate"
	ONE_IMAGE_DELETE     = "one.image.delete"
	ONE_IMAGE_PERSISTENT = "one.image.persistent"
	ONE_IMAGE_TYPECHANGE = "one.image.chtype"
	ONE_IMAGE_RENAME     = "one.image.rename"
	ONE_IMAGE_ENABLE     = "one.image.enable"
	ONE_IMAGE_REMOVE     = "one.image.delete"
	ONE_IMAGE_UPDATE     = "one.image.update"

	// virtualmachine action methods
	VM_INFO              = "one.vm.info"
	DISK_ATTACH          = "one.vm.attach"
	DISK_DETACH          = "one.vm.detach"
	ONE_VM_ACTION        = "one.vm.action"
	ONE_DISK_SNAPSHOT    = "one.vm.disksaveas"
	ONE_RECOVER          = "one.vm.recover"
	DISK_SNAPSHOT_CREATE = "one.vm.disksnapshotcreate"
	DISK_SNAPSHOT_DELETE = "one.vm.disksnapshotdelete"
	DISK_SNAPSHOT_REVERT = "one.vm.disksnapshotrevert"
	VMPOOL_ACCOUNTING    = "one.vmpool.accounting"
	VMPOOL_INFO          = "one.vmpool.info"
	ONE_VM_ATTACHNIC     = "one.vm.attachnic"
	ONE_VM_DETACHNIC     = "one.vm.detachnic"
	ONE_VM_RESIZE	     = "one.vm.resize"

	// template action methods
	TEMPLATE_INSTANTIATE  = "one.template.instantiate"
	TEMPLATEPOOL_INFO     = "one.templatepool.info"
	TEMPLATE_UPDATE       = "one.template.update"
	ONE_TEMPLATE_ALLOCATE = "one.template.allocate"
)

var (
	ErrArgsNotSatisfied = errors.New("[" + ENDPOINT + "," + USERID + "," + PASSWORD + "] one (or) more args missing!")
)

type Rpc struct {
	Client xmlrpc.Client
	Key    string
}

func NewClient(config map[string]string) (*Rpc, error) {
	log.Debugf(cmd.Colorfy("  > [one-go] connecting", "blue", "", "bold"))

	if !satisfied(config) {
		return nil, ErrArgsNotSatisfied
	}

	client, err := xmlrpc.NewClient(config[ENDPOINT], nil)
	if err != nil {
		return nil, err
	}
	//log.Debugf(cmd.Colorfy("  > [one-go] connection response", "blue", "", "bold")+"%#v",client)
	log.Debugf(cmd.Colorfy("  > [one-go] connected", "blue", "", "bold")+" %s", config[ENDPOINT])

	return &Rpc{
		Client: *client,
		Key:    config[USERID] + ":" + config[PASSWORD]}, nil
}

func (c *Rpc) Call(command string, args []interface{}) (string, error) {
	log.Debugf(cmd.Colorfy("  > [one-go] ", "blue", "", "bold")+"%s", command)
	log.Debugf(cmd.Colorfy("\n> args   ", "cyan", "", "bold")+" %v\n", args)

	result := []interface{}{}

	if err := c.Client.Call(command, args, &result); err != nil {
		return "", err
	}

	res, err := c.isSuccess(result)
	if err != nil {
		return "", err
	}
	//log.Debugf(cmd.Colorfy("\n> response ", "cyan", "", "bold")+" %v", result)
	log.Debugf(cmd.Colorfy("  > [one-go] ( ´ ▽ ` ) SUCCESS", "blue", "", "bold"))
	return res, nil
}

func (c *Rpc) isSuccess(result []interface{}) (string, error) {
	var res string
	success := result[0].(bool)
	if !success {
		return "", errors.New(result[1].(string))
	}
	if w, ok := result[1].(int64); ok {
		res = strconv.FormatInt(w, 10)
	} else if w, ok := result[1].(string); ok {
		res = w
	}
	//result[1] is error message or ID of action vm,vnet,cluster and etc.,
	return res, nil
}

func satisfied(c map[string]string) bool {
	return len(strings.TrimSpace(c[ENDPOINT])) > 0 &&
		len(strings.TrimSpace(c[USERID])) > 0 &&
		len(strings.TrimSpace(c[PASSWORD])) > 0
}
