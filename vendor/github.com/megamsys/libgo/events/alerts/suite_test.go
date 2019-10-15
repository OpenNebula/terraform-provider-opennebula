package alerts

import (
	constants "github.com/megamsys/libgo/utils"
	"gopkg.in/check.v1"
	"os"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	Meta   map[string]string
	Mailer map[string]string
}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	meta := make(map[string]string, 0)
	mailer := make(map[string]string, 0)
	home := os.Getenv("MEGAM_HOME")
	dir := "/vertice/"
	if dir != "" {
		meta[constants.HOME] = home
		dir = home + dir
	} else {
		meta[constants.HOME] = "/var/lib/megam"
		dir = "/var/lib/megam" + dir
	}
	meta[constants.DIR] = dir
	mailer[constants.ENABLED] = "true"
	mailer[constants.DOMAIN] = "smtp.mailgun.org"
	mailer[constants.USERNAME] = "test@ojamail.megambox.com"
	mailer[constants.SENDER] = "info@megam.io"
	mailer[constants.PASSWORD] = "123456789"
	mailer[constants.IDENTITY] = ""
	mailer[constants.NILAVU] = "https://console.megam.io"
	mailer[constants.LOGO] = "https://s3-ap-southeast-1.amazonaws.com/megampub/images/mailers/megam_vertice.png"
	s.Mailer = mailer
	s.Meta = meta
}
