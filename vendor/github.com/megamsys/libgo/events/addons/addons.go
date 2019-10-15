package addons

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/api"
	"github.com/megamsys/libgo/events/alerts"
	constants "github.com/megamsys/libgo/utils"
)

const (
	ADDONS_NEW    = "/addons/content"
	GETADDONS     = "/addons/"
	PROVIDER_NAME = "provider_name"
	PROVIDER_ID   = "provider_id"
)

type Addons struct {
	Id           string   `json:"id" cql:"id"`
	ProviderName string   `json:"provider_name" cql:"provider_name"`
	ProviderId   string   `json:"provider_id" cql:"provider_id"`
	AccountId    string   `json:"account_id" cql:"account_id"`
	Options      []string `json:"options" cql:"options"`
}

type ApiAddons struct {
	JsonClaz string   `json:"json_claz"`
	Results  []Addons `json:"results"`
}

func NewAddons(edata alerts.EventData) *Addons {
	return &Addons{
		Id:           "",
		ProviderName: edata.M[PROVIDER_NAME],
		ProviderId:   edata.M[PROVIDER_ID],
		AccountId:    edata.M[constants.ACCOUNT_ID],
		Options:      edata.D,
	}
}

func (s *Addons) Onboard(m map[string]string) error {
	args := api.NewArgs(m)
	cl := api.NewClient(args, ADDONS_NEW)
	_, err := cl.Post(s)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}
	return nil
}

func (s *Addons) Get(m map[string]string) (*Addons, error) {
	// Here skips balances fetching for the VMs which is launched on opennebula,
	// that does not have records on vertice database
	if s.AccountId == "" {
		return nil, fmt.Errorf("account_id should not be empty")
	}
	args := api.NewArgs(m)
	args.Email = s.AccountId
	cl := api.NewClient(args, GETADDONS+s.ProviderName)
	response, err := cl.Get()
	if err != nil {
		return nil, err
	}
	o := &ApiAddons{}
	err = json.Unmarshal(response, o)
	if err != nil {
		return nil, err
	}

	return &o.Results[0], nil
}
