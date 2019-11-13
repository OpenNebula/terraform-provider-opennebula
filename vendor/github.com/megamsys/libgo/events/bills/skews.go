/*
** Copyright [2013-2017] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */
package bills

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/api"
	"github.com/megamsys/libgo/events/alerts"
	"github.com/megamsys/libgo/pairs"
	constants "github.com/megamsys/libgo/utils"
	"strconv"
	"time"
)

const (
	EVENTSKEWS          = "/eventsskews"
	EVENTSKEWS_NEW      = "/eventsskews/content"
	EVENTSKEWS_UPDATE   = "/eventsskews/update"
	EVENTEVENTSKEWSJSON = "Megam::Skews"
	HARDSKEWS           = "terminate"
	SOFTSKEWS           = "suspend"
	WARNING             = "warning"
	ACTIVE              = "active"
	RESUME              = "start"
)

type ApiSkewsEvents struct {
	JsonClaz string         `json:"json_claz"`
	Results  []*EventsSkews `json:"results"`
}

type EventsSkews struct {
	Id        string          `json:"id"`
	AccountId string          `json:"account_id"`
	CatId     string          `json:"cat_id"`
	Inputs    pairs.JsonPairs `json:"inputs"`
	Outputs   pairs.JsonPairs `json:"outputs"`
	Actions   pairs.JsonPairs `json:"actions"`
	JsonClaz  string          `json:"json_claz"`
	Status    string          `json:"status"`
	EventType string          `json:"event_type"`
}

type updateSkews struct {
	EventsSkews
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewEventsSkews(email, cat_id string, mi map[string]string) ([]*EventsSkews, error) {

	if email == "" {
		return nil, fmt.Errorf("account_id should not be empty")
	}

	args := api.NewArgs(mi)
	args.Email = email
	cl := api.NewClient(args, EVENTSKEWS+"/"+cat_id)
	response, err := cl.Get()
	if err != nil {
		return nil, err
	}

	ac := &ApiSkewsEvents{}
	err = json.Unmarshal(response, ac)
	if err != nil {
		return nil, err
	}
	return ac.Results, nil
}

func (s *EventsSkews) update(mi map[string]string) error {
	args := api.NewArgs(mi)
	args.Email = s.AccountId
	cl := api.NewClient(args, EVENTSKEWS_UPDATE)
	_, err := cl.Post(s.updateData())
	return err
}

func (s *EventsSkews) updateData() *updateSkews {
	return &updateSkews{
		EventsSkews: *s,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (s *EventsSkews) CreateEvent(o *BillOpts, ACTION string, mi map[string]string) error {
	var exp_at, gen_at time.Time
	var action, next string
	var err error
	mm := make(map[string][]string, 0)
	if s.Inputs != nil {
		gen_at, err = time.Parse(time.RFC3339, s.Inputs.Match(constants.ACTION_TRIGGERED_AT))
		if err != nil {
			return err
		}
	} else {
		gen_at = time.Now()
	}

	softDue, err := time.ParseDuration(o.SoftGracePeriod)
	hardDue, err := time.ParseDuration(o.HardGracePeriod)
	if err != nil {
		return err
	}
	switch ACTION {
	case HARDSKEWS:
		exp_at = gen_at.Add(hardDue)
		action = HARDSKEWS
		next = "unrecoverable"
	case SOFTSKEWS:
		exp_at = gen_at.Add(hardDue)
		action = SOFTSKEWS
		next = HARDSKEWS
	case WARNING:
		mm[constants.ACTION_TRIGGERED_AT] = []string{gen_at.Format(time.RFC3339)}
		exp_at = gen_at.Add(softDue)
		action = WARNING
		next = SOFTSKEWS
	}
	mm[constants.NEXT_ACTION_DUE_AT] = []string{exp_at.Format(time.RFC3339)}
	mm[constants.ACTION] = []string{action}
	mm[constants.NEXT_ACTION] = []string{next}
	mm[constants.ASSEMBLIESID] = []string{o.AssembliesId}

	s.Inputs.NukeAndSet(mm)
	s.Status = ACTIVE
	return s.Create(mi, o)
}

func (s *EventsSkews) Create(mi map[string]string, o *BillOpts) error {
	args := api.NewArgs(mi)
	args.Email = s.AccountId
	cl := api.NewClient(args, EVENTSKEWS_NEW)
	_, err := cl.Post(s)
	if err != nil {
		return err
	}
	err = s.PushSkews(mi, s.Inputs.Match(constants.ACTION))
	if err != nil {
		return err
	}
	return s.skewsMailer(o)
}

func (sk *EventsSkews) skewsMailer(o *BillOpts) error {
	mm := make(map[string]string, 0)
	softDue, _ := time.ParseDuration(o.SoftGracePeriod)
	hardDue, _ := time.ParseDuration(o.HardGracePeriod)
	mm[constants.EMAIL] = sk.AccountId
	mm[constants.VERTNAME] = o.AssemblyName
	mm[constants.SOFT_ACTION] = SOFTSKEWS
	mm[constants.SOFT_GRACEPERIOD] = strconv.FormatInt(int64(softDue.Seconds()/3600/24), 10)
	mm[constants.SOFT_LIMIT] = o.SoftLimit
	mm[constants.HARD_GRACEPERIOD] = strconv.FormatInt(int64(hardDue.Seconds()/3600/24), 10)
	mm[constants.HARD_ACTION] = HARDSKEWS
	mm[constants.HARD_LIMIT] = o.HardLimit
	mm[constants.ACTION_TRIGGERED_AT] = sk.Inputs.Match(constants.ACTION_TRIGGERED_AT)
	mm[constants.NEXT_ACTION_DUE_AT] = sk.Inputs.Match(constants.NEXT_ACTION_DUE_AT)
	mm[constants.ACTION] = sk.Inputs.Match(constants.ACTION)
	mm[constants.NEXT_ACTION] = sk.Inputs.Match(constants.NEXT_ACTION)
	notifier := alerts.NewMailer(alerts.Mailer, alerts.Mailer)
	return notifier.Notify(alerts.SKEWS_WARNING, alerts.EventData{M: mm})
}

func (s *EventsSkews) PushSkews(mi map[string]string, skew_action string) error {
	req := api.NewRequest(s.AccountId)
	req.CatId = s.Inputs.Match(constants.ASSEMBLIESID)
	switch skew_action {
	case HARDSKEWS:
		req.Action = constants.DESTROY
		req.Category = constants.STATE
		req.CatType = "torpedo"
	case SOFTSKEWS:
		req.Action = constants.SUSPEND
		req.Category = constants.CONTROL
		req.CatType = "torpedo"
	case RESUME:
		req.Action = constants.START
		req.Category = constants.CONTROL
		req.CatType = "torpedo"
	case WARNING:
		return nil
	}
	return req.PushRequest(mi)
}

func (s *EventsSkews) DeactiveEvents(o *BillOpts, mi map[string]string) error {
	evts, err := NewEventsSkews(o.AccountId, o.AssemblyId, mi)
	if err != nil {
		return err
	}

	if len(evts) > 0 {
		for _, evt := range evts {
			if evt != nil && evt.Status == ACTIVE {
				evt.Status = "deactive"
				if err = evt.update(mi); err != nil {
					log.Debugf("checks skews actions for ondemand")
				}
				if evt.currentSkew() == SOFTSKEWS {
					return evt.PushSkews(mi, constants.START)
				}
			}
		}

	}
	return nil
}

func (s *EventsSkews) ActionEvents(o *BillOpts, currentBal string, mi map[string]string) error {
	log.Debugf("checks skews actions for ondemand")
	sk := make(map[string]*EventsSkews, 0)
	// to get skews events for that particular cat_id/ asm_id
	evts, err := NewEventsSkews(o.AccountId, o.AssemblyId, mi)
	if err != nil {
		return err
	}

	if len(evts) > 0 {
		action := evts[0].Inputs.Match(constants.ACTION)
		sk[action] = evts[0]
		if sk[action] != nil && sk[action].Status == ACTIVE {
			switch true {
			case action == HARDSKEWS && sk[HARDSKEWS].isExpired():
				return sk[HARDSKEWS].CreateEvent(o, HARDSKEWS, mi)
			case action == SOFTSKEWS && sk[SOFTSKEWS].isExpired():
				return sk[SOFTSKEWS].CreateEvent(o, HARDSKEWS, mi)
			case action == WARNING && sk[WARNING].isExpired():
				return sk[WARNING].CreateEvent(o, SOFTSKEWS, mi)
			}
			return nil
		}
	}
	action := s.action(o, currentBal)
	return s.CreateEvent(o, action, mi)
}

func (s *EventsSkews) SkewsQuotaUnpaid(o *BillOpts, mi map[string]string) error {
	log.Debugf("checks skews actions for quota (%s).", o.QuotaId)
	actions := make(map[string]string, 0)
	sk := make(map[string]*EventsSkews, 0)
	// to get skews events for that particular cat_id/ asm_id
	evts, err := NewEventsSkews(o.AccountId, o.AssemblyId, mi)
	if err != nil {
		return err
	}

	if len(evts) > 0 && evts[0].Status == ACTIVE {
		action := evts[0].Inputs.Match(constants.ACTION)
		sk[action] = evts[0]
		actions[action] = ACTIVE
		switch true {
		case actions[HARDSKEWS] == ACTIVE && sk[HARDSKEWS].isExpired():
			return sk[HARDSKEWS].CreateEvent(o, HARDSKEWS, mi)
		case actions[SOFTSKEWS] == ACTIVE && sk[SOFTSKEWS].isExpired():
			return sk[SOFTSKEWS].CreateEvent(o, HARDSKEWS, mi)
		case actions[WARNING] == ACTIVE && sk[WARNING].isExpired():
			return sk[WARNING].CreateEvent(o, SOFTSKEWS, mi)
		}
		return nil
	}
	return s.CreateEvent(o, WARNING, mi)
}

func (s *EventsSkews) action(o *BillOpts, currentBal string) string {
	cb, _ := strconv.ParseFloat(currentBal, 64)
	slimit, _ := strconv.ParseFloat(o.SoftLimit, 64)
	hlimit, _ := strconv.ParseFloat(o.HardLimit, 64)
	if cb <= hlimit {
		return HARDSKEWS
	} else if cb <= slimit {
		return SOFTSKEWS
	}
	return WARNING
}

func (s *EventsSkews) isExpired() bool {
	t1, _ := time.Parse(time.RFC3339, s.Inputs.Match(constants.ACTION_TRIGGERED_AT))
	t2, _ := time.Parse(time.RFC3339, s.Inputs.Match(constants.NEXT_ACTION_DUE_AT))
	duration := t2.Sub(t1)
	return t1.Add(duration).Sub(time.Now()) < time.Minute
}

func (s *EventsSkews) currentSkew() string {
	return s.Inputs.Match(constants.ACTION)
}
