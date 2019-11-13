package metrics

/*
import (
	"encoding/xml"
	"testing"
	"time"
	"github.com/megamsys/opennebula-go/api"
	"gopkg.in/check.v1"
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
	cm[api.ENDPOINT] = "http://192.168.0.118:2633/RPC2"
	cm[api.USERID] = "oneadmin"
	cm[api.PASSWORD] = "WuedmopFupt6"
	s.cm = cm
}

func (s *S) TestGet(c *check.C) {
	client, _ := api.NewClient(s.cm)
	vm := Accounting{Api: client, StartTime: time.Now().Add(-10 * time.Minute).Unix(), EndTime: time.Now().Unix()}
	res, error := vm.Get()
	c.Assert(error, check.IsNil)
  se := []byte(res.(string))
	ons := &OpenNebulaStatus{}
	error = xml.Unmarshal(se, ons)
	c.Assert(error, check.IsNil)
}
// */
