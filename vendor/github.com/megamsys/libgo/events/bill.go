package events

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/events/alerts"
	"github.com/megamsys/libgo/events/bills"
	_ "github.com/megamsys/libgo/events/bills"
	constants "github.com/megamsys/libgo/utils"
	"reflect"
	"strings"
)

type Bill struct {
	piggyBanks string
	stop       chan struct{}
	M          map[string]string
}

func NewBill(b map[string]string, m map[string]string) *Bill {
	MapCopy(m, b)
	return &Bill{
		piggyBanks: b[constants.PIGGYBANKS],
		M:          m,
	}
}

// Watches for new vms, or vms destroyed.
func (self *Bill) Watch(eventsChannel *EventChannel) error {
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
				case event.EventAction == alerts.INSUFFICIENT_FUND:
					err := self.insufficientFund(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.DEDUCT && self.M[constants.ENABLED] == constants.TRUE:
					err := self.deduct(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.BILLEDHISTORY:
					err := self.billedhistory(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.TRANSACTION:
					err := self.transaction(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.SKEWS_ACTIONS || event.EventAction == alerts.QUOTA_UNPAID:
					err := self.skews(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				}
			case <-self.stop:
				log.Info("bill watcher exiting")
				return
			}
		}
	}()
	return nil
}

func (self *Bill) skip(k string) bool {
	return !strings.Contains(self.piggyBanks, k)
}

func (self *Bill) insufficientFund(evt *Event) error {
	var err error
	a := notifiers[constants.SMTP]
	err = a.Notify(evt.EventAction, evt.EventData)
	if err != nil {
		return err
	}
	return nil
}

func (self *Bill) OnboardFunc(evt *Event) error {
	return nil
}

func (self *Bill) deduct(evt *Event) error {
	log.Infof("Event:BILL:deduct")
	result := &bills.BillOpts{}
	_ = result.FillStruct(evt.EventData.M) //we will manage error later
	for k, bp := range bills.BillProviders {
		if !self.skip(k) {
			err := bp.Deduct(result, self.M)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *Bill) skews(evt *Event) error {
	log.Infof("Event:BILL:skewsCheck")

	result := &bills.BillOpts{}
	_ = result.FillStruct(evt.EventData.M) //we will manage error later
	if !self.skip(constants.SCYLLAMGR) {
		p := bills.BillProviders[constants.SCYLLAMGR]
		if evt.EventAction == alerts.SKEWS_ACTIONS || evt.EventAction == alerts.QUOTA_UNPAID {
			return p.AuditUnpaid(result, self.M)
		}
	}
	return nil
}

func (self *Bill) billedhistory(evt *Event) error {
	log.Infof("Event:BILL:transaction")
	result := &bills.BillOpts{}
	_ = result.FillStruct(evt.EventData.M) //we will manage error later
	if !self.skip(constants.SCYLLAMGR) {
		err := bills.BillProviders[constants.SCYLLAMGR].Transaction(result, self.M)
		if err != nil {
			return err
		}
	}
	return nil
}

//transaction for all managers like whmcs and scylla
func (self *Bill) transaction(evt *Event) error {
	log.Infof("Event:BILL:transaction")
	result := &bills.BillOpts{}
	_ = result.FillStruct(evt.EventData.M) //we will manage error later
	for k, bp := range bills.BillProviders {
		if !self.skip(k) {
			err := bp.Transaction(result, self.M)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *Bill) Close() {
	if self.stop != nil {
		close(self.stop)
	}
}

func MapCopy(dst, src interface{}) {
	dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)

	for _, k := range sv.MapKeys() {
		dv.SetMapIndex(k, sv.MapIndex(k))
	}
}
