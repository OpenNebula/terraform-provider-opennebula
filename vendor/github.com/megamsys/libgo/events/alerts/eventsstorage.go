package alerts

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/api"
	"github.com/megamsys/libgo/pairs"
	constants "github.com/megamsys/libgo/utils"
)

const EVENTSTORAGE_NEW = "/eventsstorage/content"

type EventsStorage struct {
	EventType string          `json:"event_type" cql:"event_type"`
	AccountId string          `json:"account_id" cql:"account_id"`
	Data      pairs.JsonPairs `json:"data" cql:"data"`
}

func (v *VerticeApi) NotifyStorage(eva EventAction, edata EventData) error {

	if !v.satisfied(eva) {
		return nil
	}
	sdata := parseMapToOutputFormat(edata)
	v.Args.Email = edata.M[constants.ACCOUNT_ID]
	cl := api.NewClient(v.Args, EVENTSTORAGE_NEW)
	_, err := cl.Post(sdata)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}
	return nil
}

func parseMapToOutputStorage(edata EventData) EventsStorage {
	return EventsStorage{
		EventType: edata.M[constants.EVENT_TYPE],
		AccountId: edata.M[constants.ACCOUNT_ID],
		Data:      *pairs.ArrayToJsonPairs(edata.D),
	}
}
