package events

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/events/alerts"
	constants "github.com/megamsys/libgo/utils"
)

type Machine struct {
	stop chan struct{}
	fns  AfterFuncsMap
}

// Watches for new vms, or vms destroyed.
func (self *Machine) Watch(eventsChannel *EventChannel) error {
	self.stop = make(chan struct{})
	go func() {
		for {
			select {
			case event := <-eventsChannel.channel:
				switch {
				case event.EventAction == alerts.LAUNCHED:
					err := self.create(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.RUNNING:
					err := self.alert(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.DESTROYED:
					err := self.destroy()
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.SNAPSHOTTING:
					err := self.snapcreate(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.INSUFFICIENT_FUND:
					err := self.insufficientFund(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.SNAPSHOTTED:
					err := self.snapdone(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.FAILURE:
					err := self.alert(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				}
			case <-self.stop:
				log.Info("machine watcher exiting")
				return
			}
		}
	}()
	return nil
}

func (self *Machine) Close() {
	if self.stop != nil {
		close(self.stop)
	}
}

func (self *Machine) create(evt *Event) error {
	return nil
}

func (self *Machine) destroy() error {
	return nil
}

func (self *Machine) snapcreate(evt *Event) error {
	return nil
}

func (self *Machine) snapdone(evt *Event) error {
	return nil
}

func (self *Machine) insufficientFund(evt *Event) error {
	var err error
	a := notifiers[constants.SMTP]
	err = a.Notify(evt.EventAction, evt.EventData)
	if err != nil {
		return err
	}
	return nil
}

func (self *Machine) alert(evt *Event) error {
	var err error
	for _, a := range notifiers {
		err = a.Notify(evt.EventAction, evt.EventData)
	}
	if err != nil {
		return err
	}
	return self.after(evt)
}

func (self *Machine) after(evt *Event) error {
	var err error
	perActionfns := self.fns[evt.EventAction]
	for _, fn := range perActionfns {
		err = fn(evt)
	}
	return err
}
