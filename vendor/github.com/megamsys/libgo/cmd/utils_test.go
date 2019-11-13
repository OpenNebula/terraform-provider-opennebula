package cmd

import (
	"os"
	"path"

	"gopkg.in/check.v1"
	"launchpad.net/gnuflag"
)

func (s *S) TestJoinWithUserDir(c *check.C) {
	expected := path.Join(os.Getenv("MEGAM_HOME"), "a", "b")
	path := JoinWithUserDir("a", "b")
	c.Assert(path, check.Equals, expected)
}

func (s *S) TestJoinWithUserDirHomePath(c *check.C) {
	defer os.Setenv("MEGAM_HOME", os.Getenv("MEGAM_HOME"))
	os.Setenv("MEGAM_HOME", "")
	os.Setenv("MEGAM_HOME", "/wat")
	path := JoinWithUserDir("a", "b")
	c.Assert(path, check.Equals, "/wat/a/b")
}

func (s *S) TestMergeFlagSet(c *check.C) {
	var x, y bool
	fs1 := gnuflag.NewFlagSet("x", gnuflag.ExitOnError)
	fs1.BoolVar(&x, "x", false, "Something")
	fs2 := gnuflag.NewFlagSet("y", gnuflag.ExitOnError)
	fs2.BoolVar(&y, "y", false, "Something")
	ret := MergeFlagSet(fs1, fs2)
	c.Assert(ret, check.Equals, fs1)
	fs1.Parse(true, []string{"-x", "-y"})
	c.Assert(x, check.Equals, true)
	c.Assert(y, check.Equals, true)
}
