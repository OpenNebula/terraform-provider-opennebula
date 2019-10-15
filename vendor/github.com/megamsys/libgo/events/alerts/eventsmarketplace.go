package alerts

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/api"
	"github.com/megamsys/libgo/pairs"
	constants "github.com/megamsys/libgo/utils"
)

const (
	EVENTSMARKETPLACES_NEW = "/eventsmarketplace/content"
)

type EventsMarketplace struct {
	EventType     string          `json:"event_type" cql:"event_type"`
	AccountId     string          `json:"account_id" cql:"account_id"`
	MarketplaceId string          `json:"marketplace_id" cql:"marketplace_id"`
	Data          pairs.JsonPairs `json:"data" cql:"data"`
}

func (v *VerticeApi) NotifyMarketplace(eva EventAction, edata EventData) error {
	if !v.satisfied(eva) {
		return nil
	}
	sdata := parseMapToOutputMarket(edata)
	v.Args.Email = edata.M[constants.ACCOUNT_ID]
	cl := api.NewClient(v.Args, EVENTSMARKETPLACES_NEW)
	_, err := cl.Post(sdata)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}
	return nil
}

func parseMapToOutputMarket(edata EventData) EventsMarketplace {
	return EventsMarketplace{
		EventType:     edata.M[constants.EVENT_TYPE],
		AccountId:     edata.M[constants.ACCOUNT_ID],
		MarketplaceId: edata.M[constants.MARKETPLACE_ID],
		Data:          *pairs.ArrayToJsonPairs(edata.D),
	}
}
