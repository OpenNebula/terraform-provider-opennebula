package errors

import (
	"gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestValidationError(c *check.C) {
	e := ValidationError{Message: "something"}
	c.Assert(e.Error(), check.Equals, "something")
}
