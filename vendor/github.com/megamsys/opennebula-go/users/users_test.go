package users

import (
	"github.com/megamsys/opennebula-go/api"
	"gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type S struct {
	cm map[string]string
}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	cm := make(map[string]string)
	cm[api.ENDPOINT] = "http://localhost:2633/RPC2"
	cm[api.USERID] = "oneadmin"
	cm[api.PASSWORD] = "oneadmin"
	s.cm = cm
}

func (s *S) TestGetUsers(c *check.C) {
	client, _ := api.NewClient(s.cm)
	u := User{
		UserName:   "vijaym@megam.io",
		Password:   "team4megam",
		AuthDriver: "core",
		GroupIds:   []int{0},
	}
	vm := UserTemplate{
		T:     client,
		Users: u,
	}
	_, err := vm.CreateUsers()
	c.Assert(err, check.NotNil)
}
