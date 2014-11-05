package core

import (
	. "gopkg.in/check.v1"
)

func (_ *CoreSuite) TestDockerfile_Get(c *C) {
	d := NewDockerfile(
		[]byte("$DOCKERSHIP_PROJECT/$DOCKERSHIP_ENV/$DOCKERSHIP_VCS/$DOCKERSHIP_REV"),
		&Project{Name: "foo", Repository: "qux"},
		Revision{"foo": "baz"},
		&Environment{Name: "bar"},
	)

	c.Assert(string(d.Get()), Equals, "foo/bar/qux/baz")
}
