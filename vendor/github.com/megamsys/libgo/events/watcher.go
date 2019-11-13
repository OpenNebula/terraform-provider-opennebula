package events

import (
	constants "github.com/megamsys/libgo/utils"
	"time"
)

var eventTypes = map[string]EventType{
	"machine":   constants.EventMachine,
	"container": constants.EventContainer,
	"bill":      constants.EventBill,
	"user":      constants.EventUser,
}

// Interface for event  operation handlers.
type Watcher interface {
	Watch(eventChannel *EventChannel) error
}

type eventReqOpts struct {
	etype     EventType
	maxEvents int
	startTime string
	endTime   string
}

func getEventRequest(o *eventReqOpts) (*Request, bool, error) {
	ereq := NewRequest(o)
	stream := false

	ereq.EventType[o.etype] = true

	if o.maxEvents > 0 {
		ereq.maxEventsReturned = o.maxEvents
	}

	if len(o.startTime) > 0 {
		newTime, err := time.Parse(time.RFC3339, o.startTime)
		if err == nil {
			ereq.StartTime = newTime
		}
	}

	if len(o.endTime) > 0 {
		newTime, err := time.Parse(time.RFC3339, o.endTime)
		if err == nil {
			ereq.EndTime = newTime
		}
	}
	return ereq, stream, nil
}
