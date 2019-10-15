package os

import (
	"runtime"

	gc "gopkg.in/check.v1"
)

type osSuite struct {
}

var _ = gc.Suite(&osSuite{})

func (s *osSuite) TestHostOS(c *gc.C) {
	os := HostOS()
	switch runtime.GOOS {
	case "windows":
		c.Assert(os, gc.Equals, Windows)
	case "darwin":
		c.Assert(os, gc.Equals, OSX)
	case "linux":
		if os != Ubuntu && os != CentOS && os != Arch && os != Debian {
			c.Fatalf("unknown linux version: %v", os)
		}
	case "freebsd":
		c.Assert(os, gc.Equals, FreeBSD)
	default:
		c.Fatalf("unsupported operating system: %v", runtime.GOOS)
	}
}
