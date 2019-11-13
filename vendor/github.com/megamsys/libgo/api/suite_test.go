package api

import (
	"gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	ApiArgs ApiArgs
}

var _ = check.Suite(&S{})

//we need make sure the stub deploy methods are supported.
func (s *S) SetUpSuite(c *check.C) {
	s.ApiArgs = ApiArgs{
		Email:      "info@megam.io",
		Url:        "http://192.168.0.1:9000/v2",
		Api_Key:    "",
		Master_Key: "3b8eb672aa7c8db82e5d34a0744740b20ed59e1f6814cfb63364040b0994ee3f",
		Password:   "",
		Org_Id:     "",
	}
}

// func (s *S) TearDownSuite(c *check.C) {
//   //just stop the server here.
// }
