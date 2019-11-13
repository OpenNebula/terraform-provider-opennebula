package api

import (
	"fmt"
	"gopkg.in/check.v1"
)

func (s *S) TestBindService(c *check.C) {
	err := fmt.Errorf("error")
	c.Assert(err, check.NotNil)
}
