package events

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/events/alerts"
)

type Container struct {
	stop chan struct{}
}

// Watches for new vms, or vms destroyed.
func (self *Container) Watch(eventsChannel *EventChannel) error {
	self.stop = make(chan struct{})
	go func() {
		for {
			select {
			case event := <-eventsChannel.channel:
				switch {
				case event.EventAction == alerts.LAUNCHED:
					err := self.create()
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.DESTROYED:
					err := self.destroy()
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				}
			case <-self.stop:
				log.Info("container watcher exiting")
				return
			}
		}
	}()
	return nil
}

func (self *Container) Close() {
	if self.stop != nil {
		close(self.stop)
	}
}

func (self *Container) create() error {
	log.Info("RECV container create")
	return nil
}

func (self *Container) destroy() error {
	log.Info("RECV container destroy")
	return nil
}
