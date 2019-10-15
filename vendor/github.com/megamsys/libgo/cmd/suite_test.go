package cmd

import (
	"bytes"
	"os"
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	stdin        *os.File
	recover      []string
	recoverToken []string
}

var _ = check.Suite(&S{})
var manager *Manager

func (s *S) SetUpTest(c *check.C) {
	var stdout, stderr bytes.Buffer
	manager = NewManager("megamd", "1.0", &stdout, &stderr, os.Stdin, nil, nil)
	var exiter recordingExiter
	manager.e = &exiter
	os.Setenv("MEGAM_HOME", "/home/megam")
}

func (s *S) TearDownTest(c *check.C) {
	os.Unsetenv("MEGAM_HOME")
}
