package alerts

import (
	constants "github.com/megamsys/libgo/utils"
)

type infobip struct {
	url            string
	username       string
	password       string
	api_key        string
	application_id string
	message_id     string
}

func NewInfobip(m map[string]string) Notifier {
	return &infobip{
		url:            "https://infobip.com/v2",
		username:       m[constants.USERNAME],
		password:       m[constants.PASSWORD],
		api_key:        m[constants.API_KEY],
		application_id: m[constants.APPLICATION_ID],
		message_id:     m[constants.MESSAGE_ID],
	}
}

func (i *infobip) satisfied(eva EventAction) bool {
	if eva == STATUS {
		return false
	}
	return true
}

func (i *infobip) Notify(eva EventAction, edata EventData) error {
	return nil
}
