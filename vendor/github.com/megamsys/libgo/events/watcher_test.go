package events

import (
	"testing"
	"time"

	"github.com/megamsys/libgo/events/alerts"
	constants "github.com/megamsys/libgo/utils"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	fakeEvent1    *Event
	fakeEvent2    *Event
	myEventHolder *events
	myRequest     *Request
}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	fakeEvent1 := makeEvent(createOldTime(), constants.EventBill, alerts.ONBOARD)
	fakeEvent2 := makeEvent(time.Now(), constants.EventBill, alerts.DEDUCT)
	c.Assert(fakeEvent1, check.NotNil)
	c.Assert(fakeEvent2, check.NotNil)

	s.myEventHolder = NewEventManager(DefaultStoragePolicy())
	req, _, err := getEventRequest(&eventReqOpts{
		etype: constants.EventBill,
	})
	c.Assert(err, check.IsNil)
	s.myRequest = req
	s.fakeEvent1 = fakeEvent1
	s.fakeEvent2 = fakeEvent2
	c.Assert(s.myRequest.StartTime.IsZero(), check.Equals, true)
	c.Assert(s.myRequest.EndTime.IsZero(), check.Equals, true)
}

func makeEvent(inTime time.Time, evtType EventType, evtAction alerts.EventAction) *Event {
	return &Event{
		Timestamp:   inTime,
		EventAction: evtAction,
		EventType:   evtType,
	}
}

func createOldTime() time.Time {
	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	linetime, err := time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	if err != nil {
		//fmt.Errorf("could not format time.Time object")
	} else {
		return linetime
	}
	return time.Now()
}

func (s *S) TestAddEventAddsEventsToEventManager(c *check.C) {
	s.myEventHolder.AddEvent(s.fakeEvent1)
	c.Assert(1, check.Equals, len(s.myEventHolder.eventStore))
	c.Assert(s.fakeEvent1, check.DeepEquals, s.myEventHolder.eventStore[constants.EventBill].Get(0).(*Event))
}

func (s *S) TestWatchEventsDetectsNewEvents(c *check.C) {
	s.myRequest.StartTime = time.Time{}
	s.myRequest.EndTime = time.Time{}
	c.Assert(s.myRequest.StartTime.IsZero(), check.Equals, true)
	c.Assert(s.myRequest.EndTime.IsZero(), check.Equals, true)

	returnEventChannel, err := s.myEventHolder.WatchEvents(s.myRequest)
	c.Assert(err, check.IsNil)

	s.myEventHolder.AddEvent(s.fakeEvent1)
	s.myEventHolder.AddEvent(s.fakeEvent2)

	startTime := time.Now()
	go func() {
		time.Sleep(5 * time.Second)
		if time.Since(startTime) > (5 * time.Second) {
			//fmt.Errorf("Took too long to receive all the events")
		}
	}()

	eventsFound := 0
	go func() {
		for event := range returnEventChannel.GetChannel() {
			eventsFound += 1
			if eventsFound == 1 {
				c.Assert(s.fakeEvent1.EventAction, check.DeepEquals, event.EventAction)
			} else if eventsFound == 2 {
				c.Assert(s.fakeEvent2.EventAction, check.DeepEquals, event.EventAction)
				break
			}
		}
	}()
}

func (s *S) TestGetEventsForOneEvent(c *check.C) {
	s.myRequest.maxEventsReturned = 1
	s.myEventHolder.AddEvent(s.fakeEvent1)
	s.myEventHolder.AddEvent(s.fakeEvent2)

	receivedEvents, err := s.myEventHolder.GetEvents(s.myRequest)
	c.Assert(err, check.IsNil)
	c.Assert(1, check.Equals, len(receivedEvents))
	c.Assert(s.fakeEvent2, check.DeepEquals, receivedEvents[0])
}

func (s *S) TestGetEventsForTimePeriod(c *check.C) {
	s.myRequest.StartTime = time.Now().Add(-1 * time.Second * 10)
	s.myRequest.EndTime = time.Now().Add(time.Second * 10)

	s.myEventHolder.AddEvent(s.fakeEvent1)
	s.myEventHolder.AddEvent(s.fakeEvent2)

	receivedEvents, err := s.myEventHolder.GetEvents(s.myRequest)
	c.Assert(err, check.IsNil)

	c.Assert(1, check.Equals, len(receivedEvents))
	c.Assert(s.fakeEvent2, check.DeepEquals, receivedEvents[0])
}

func (s *S) TestGetEventsForNoTypeRequested(c *check.C) {
	s.myEventHolder.AddEvent(s.fakeEvent1)
	s.myEventHolder.AddEvent(s.fakeEvent2)

	receivedEvents, err := s.myEventHolder.GetEvents(s.myRequest)
	c.Assert(err, check.IsNil)
	c.Assert(1, check.Equals, len(receivedEvents))
}
