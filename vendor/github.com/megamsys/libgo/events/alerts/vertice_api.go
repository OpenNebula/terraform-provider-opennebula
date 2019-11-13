package alerts

import (
	"github.com/megamsys/libgo/api"
	constants "github.com/megamsys/libgo/utils"
	"strings"
)

type VerticeApi struct {
	Args api.ApiArgs
}

func NewApiArgs(args map[string]string) Notifier {
	return &VerticeApi{
		Args: api.ApiArgs{
			Email:      args[constants.USERMAIL],
			Api_Key:    args[constants.API_KEY],
			Master_Key: args[constants.MASTER_KEY],
			Password:   args[constants.PASSWORD],
			Url:        args[constants.API_URL],
		},
	}
}

func (v *VerticeApi) satisfied(eva EventAction) bool {
	if eva == STATUS {
		return true
	}
	return false
}

func (s *VerticeApi) Notify(eva EventAction, edata EventData) error {
	value := edata.M[constants.EVENT_TYPE]
	et := strings.Split(value, ".")
	if et[0] == "compute" && et[1] == "instance" {
		return s.NotifyVm(eva, edata)
	} else if et[0] == "bill" {
		return s.NotifyBill(eva, edata)
	} else if et[0] == "storage" {
		return s.NotifyStorage(eva, edata)
	} else if et[0] == "compute" && et[1] == "container" {
		return s.NotifyContainer(eva, edata)
	} else if et[0] == "marketplaces" {
		return s.NotifyMarketplace(eva, edata)
	}
	return nil
}
