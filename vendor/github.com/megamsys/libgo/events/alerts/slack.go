package alerts

import (
	"github.com/Bowery/slack"
	constants "github.com/megamsys/libgo/utils"
)

type slacker struct {
	token string
	chnl  string
}

func NewSlack(m map[string]string) Notifier {
	return &slacker{token: m[constants.TOKEN], chnl: m[constants.CHANNEL]}
}

func (s *slacker) satisfied(eva EventAction) bool {
	if eva == STATUS {
		return false
	}
	return true
}

func (s *slacker) Notify(eva EventAction, edata EventData) error {
	if !s.satisfied(eva) {
		return nil
	}
	if err := slack.NewClient(s.token).SendMessage("#"+s.chnl, edata.M["message"], "megamio"); err != nil {
		return err
	}
	return nil
}
