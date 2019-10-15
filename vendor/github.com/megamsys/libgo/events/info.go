package events

import (
	"encoding/json"
	"github.com/megamsys/libgo/events/alerts"
	"github.com/megamsys/libgo/pairs"
	constants "github.com/megamsys/libgo/utils"
	"strconv"
	"time"
)

const (
	timePrecision time.Duration = 10 * time.Millisecond // 10ms, i.e. 0.01s
)

type StoredEvent struct {
	Id         string          `json:"id"`
	AccountsId string          `json:"AccountsId"`
	Type       string          `json:"type"`
	Action     string          `json:"action"`
	Inputs     pairs.JsonPairs `json:"inputs"`
	CreatedAt  string          `json:"created_at"`
}

func NewParseEvent(b []byte) (*StoredEvent, error) {
	st := &StoredEvent{}
	err := json.Unmarshal(b, &st)
	if err != nil {
		return nil, err
	}
	return st, err
}

func (st *StoredEvent) AsEvent() (*Event, error) {
	ea, err := strconv.Atoi(st.Action)
	if err != nil {
		return nil, err
	}

	e := Event{
		AccountsId:  st.AccountsId,
		EventType:   toEventType(st.Type),
		EventAction: toEventAction(ea),
		EventData:   alerts.EventData{M: st.Inputs.ToMap()},
		Timestamp:   time.Now().Local(),
	}

	if err != nil {
		return nil, err
	}
	return &e, err
}

// Event contains information general to events such as the time at which they
// occurred, their specific type, and the actual event. Event types are
// differentiated by the EventType field of Event.
type Event struct {
	AccountsId string
	// the time at which the event occurred
	Timestamp time.Time

	// the type of event. EventType is an enumerated type
	EventType EventType

	//the action can be
	//bill create, bill delete
	EventAction alerts.EventAction

	// the original event object and all of its extraneous data, ex. an
	// OomInstance
	EventData alerts.EventData
}

func (e *Event) String() string {
	return ""
}

// EventType is an enumerated type which lists the categories under which
// events may fall. The Event field EventType is populated by this enum.
type EventType string

type EventChannel struct {
	// Watch ID. Can be used by the caller to request cancellation of watch events.
	watchId int
	// Channel on which the caller can receive watch events.
	channel chan *Event
}

func toEventAction(a int) alerts.EventAction {
	return alerts.EventAction(a)
}

func toEventType(t string) EventType {
	switch t {
	case "machine":
		return constants.EventMachine
	case "container":
		return constants.EventContainer
	case "bill":
		return constants.EventBill
	case "user":
		return constants.EventUser
	case "status":
		return constants.EventStatus
	default:
		return ""
	}
}

type MultiEvent struct {
	Events []*Event
}

func (e *MultiEvent) String() string {
	return ""
}

func NewMulti(ea []*Event) *MultiEvent {
	return &MultiEvent{Events: ea}
}

func (me *MultiEvent) Write() error {
	var err error
	me.Events = append(me.Events, &Event{}) //add the usernotification event
	for _, e := range me.Events {
		if me.IsEnabled() {
			err = W.Write(e)
		}
	}
	return err
}

func (me *MultiEvent) IsEnabled() bool {
	return W != nil
}
