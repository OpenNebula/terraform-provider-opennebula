package bills

import (
/*"fmt"
"os"
"testing"

"gopkg.in/check.v1"*/
)

/*func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	//	if aws_acc == "" || aws_sec == "" {
	c.Skip("-R53 (aws access/secret keys) not provided")
	//	}
	cf := dns.NewConfig()
	cf.AccessKey = os.Getenv("AWS_ACCESS_KEY")
	cf.SecretKey = os.Getenv("AWS_SECRET_KEY")
	s.cf = cf
}

func (s *S) TestShouldBeRegistered(c *check.C) {
	s.cf.MkGlobal()
	router.Register("route53", createRouter)
	got, err := router.Get("route53")
	c.Assert(err, check.IsNil)
	_, ok := got.(route53Router)
	c.Assert(ok, check.Equals, true)
}

func (s *S) TestOnboard(c *check.C) {
	s.cf.MkGlobal()
	vRouter, err := router.Get("route53")
	c.Assert(err, check.IsNil)
	err = vRouter.SetCName("myapp1.megambox.com", "192.168.1.100")
	c.Assert(err, check.IsNil)
}
*/
