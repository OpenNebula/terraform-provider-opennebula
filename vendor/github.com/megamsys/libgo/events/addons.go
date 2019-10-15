package events

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/events/addons"
	"github.com/megamsys/libgo/events/alerts"
)

type Addons struct {
	stop chan struct{}
	M    map[string]string
}

func NewAddons(b map[string]string, m map[string]string) *Addons {
	return &Addons{
		M: m,
	}
}

// Watches for new vms, or vms destroyed.
func (self *Addons) Watch(eventsChannel *EventChannel) error {
	self.stop = make(chan struct{})
	go func() {
		for {
			select {
			case event := <-eventsChannel.channel:
				switch {
				case event.EventAction == alerts.ONBOARD:
					err := self.OnboardFunc(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				}
			case <-self.stop:
				log.Info("addons watcher exiting")
				return
			}
		}
	}()
	return nil
}

func (self *Addons) Close() {
	if self.stop != nil {
		close(self.stop)
	}
}

func (self *Addons) OnboardFunc(evt *Event) error {
	log.Info("RECV addons onboarded")
	add := addons.NewAddons(evt.EventData)
	err := add.Onboard(self.M)
	if err != nil {
		return err
	}
	return nil
}

func (self *Addons) destroy() error {
	log.Info("RECV container destroy")
	return nil
}
