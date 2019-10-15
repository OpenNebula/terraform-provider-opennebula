package alerts

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/api"
	"github.com/megamsys/libgo/pairs"
	constants "github.com/megamsys/libgo/utils"
)

const (
	EVENTSVM_NEW = "/eventsvm/content"
)

type EventsVm struct {
	EventType  string          `json:"event_type" cql:"event_type"`
	AccountId  string          `json:"account_id" cql:"account_id"`
	AssemblyId string          `json:"assembly_id" cql:"assembly_id"`
	Data       pairs.JsonPairs `json:"data" cql:"data"`
}

func (v *VerticeApi) NotifyVm(eva EventAction, edata EventData) error {
	if !v.satisfied(eva) {
		return nil
	}
	sdata := parseMapToOutputFormat(edata)
	v.Args.Email = edata.M[constants.ACCOUNT_ID]
	cl := api.NewClient(v.Args, EVENTSVM_NEW)
	_, err := cl.Post(sdata)
	if err != nil {
		log.Debugf(err.Error())
		return err
	}
	return nil
}

func parseMapToOutputFormat(edata EventData) EventsVm {
	return EventsVm{
		EventType:  edata.M[constants.EVENT_TYPE],
		AccountId:  edata.M[constants.ACCOUNT_ID],
		AssemblyId: edata.M[constants.ASSEMBLY_ID],
		Data:       *pairs.ArrayToJsonPairs(edata.D),
	}
}
