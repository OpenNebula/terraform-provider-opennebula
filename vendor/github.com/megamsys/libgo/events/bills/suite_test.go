package bills

import (
	"github.com/megamsys/libgo/utils"
	"gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	m map[string]string
}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	m := make(map[string]string, 0)
	m[utils.MASTER_KEY] = "4ee3f"
	m[utils.API_URL] = "http://192.168.0.100:9000/v2"
	s.m = m
}
