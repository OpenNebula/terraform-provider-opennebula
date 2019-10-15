package alerts

import (
	constants "github.com/megamsys/libgo/utils"
	"gopkg.in/check.v1"
	"os"
)

var st = os.Getenv("NIL_SLACK_TOKEN")
var ch = "ahoy"

func (s *S) TestSlack(c *check.C) {
	if st == "" {
		c.Skip("-Slack (token) not provided")
	}
	c.Assert(len(st) > 0, check.Equals, true)
	ms := NewSlack(map[string]string{constants.TOKEN: st, constants.CHANNEL: ch})
	c.Assert(ms, check.NotNil)
	err := ms.Notify(LAUNCHED, EventData{M: map[string]string{"message": "Awesome vertice... :)"}})
	c.Assert(err, check.IsNil)
}
