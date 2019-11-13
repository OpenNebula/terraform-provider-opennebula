package alerts

import (
	log "github.com/Sirupsen/logrus"
	ldb "github.com/megamsys/libgo/db"
	constants "github.com/megamsys/libgo/utils/obc"
	"github.com/pborman/uuid"
	"time"
)

const EVENTSOBCBUCKET = "events_for_obc"

type ObcEvents *[]EventsObc

type EventsObc struct {
	Id        string    `json:"id" cql:"id"`
	EventType string    `json:"event_type" cql:"event_type"`
	AccountId string    `json:"account_id" cql:"account_id"`
	HostIp    string    `json:"host_ip" cql:"host_ip"`
	HostId    string    `json:"host_id" cql:"host_id"`
	Data      []string  `json:"data" cql:"data"`
	CreatedAt time.Time `json:"created_at" cql:"created_at"`
}

func (s *Scylla) NotifyOBC(eva EventAction, edata EventData) error {
	if !s.satisfied(eva) {
		return nil
	}
	s_data := parseMapToOutputObc(edata)
	ops := ldb.Options{
		TableName:   EVENTSOBCBUCKET,
		Pks:         []string{constants.EVENT_TYPE, constants.CREATED_AT},
		Ccms:        []string{constants.HOST_ID, constants.ACCOUNT_ID},
		Hosts:       s.Scylla_host,
		Keyspace:    s.Scylla_keyspace,
		Username:    s.Scylla_username,
		Password:    s.Scylla_password,
		PksClauses:  map[string]interface{}{constants.EVENT_TYPE: edata.M[constants.EVENT_TYPE], constants.CREATED_AT: s_data.CreatedAt},
		CcmsClauses: map[string]interface{}{constants.HOST_ID: edata.M[constants.HOST_ID], constants.ACCOUNT_ID: edata.M[constants.ACCOUNT_ID]},
	}
	if err := ldb.Storedb(ops, s_data); err != nil {
		log.Debugf(err.Error())
		return err
	}
	return nil
}

func (s *Scylla) GetEventsByEmail(email string, limit int) (*[]EventsObc, error) {
	events := &[]EventsObc{}
	e := EventsObc{}

	ops := ldb.Options{
		TableName:   EVENTSOBCBUCKET,
		Pks:         []string{constants.ACCOUNT_ID},
		Ccms:        []string{},
		Hosts:       s.Scylla_host,
		Keyspace:    s.Scylla_keyspace,
		Username:    s.Scylla_username,
		Password:    s.Scylla_password,
		PksClauses:  map[string]interface{}{constants.ACCOUNT_ID: email},
		CcmsClauses: map[string]interface{}{},
	}
	if err := ldb.FetchListdb(ops, limit, e, events); err != nil {
		log.Debugf(err.Error())
		return nil, err
	}
	return events, nil
}

func (s *Scylla) GetEventsByNodeId(email, id string, limit int) (*[]EventsObc, error) {
	events := &[]EventsObc{}
	e := EventsObc{}

	ops := ldb.Options{
		TableName:   EVENTSOBCBUCKET,
		Pks:         []string{constants.HOST_ID},
		Ccms:        []string{constants.ACCOUNT_ID},
		Hosts:       s.Scylla_host,
		Keyspace:    s.Scylla_keyspace,
		Username:    s.Scylla_username,
		Password:    s.Scylla_password,
		PksClauses:  map[string]interface{}{constants.HOST_ID: id},
		CcmsClauses: map[string]interface{}{constants.ACCOUNT_ID: email},
	}
	if err := ldb.FetchListdb(ops, limit, e, events); err != nil {
		log.Debugf(err.Error())
		return nil, err
	}
	return events, nil
}

func parseMapToOutputObc(edata EventData) EventsObc {
	return EventsObc{
		Id:        uuid.New(),
		EventType: edata.M[constants.EVENT_TYPE],
		AccountId: edata.M[constants.ACCOUNT_ID],
		HostIp:    edata.M[constants.HOST_IP],
		HostId:    edata.M[constants.HOST_ID],
		Data:      edata.D,
		CreatedAt: time.Now(),
	}
}
