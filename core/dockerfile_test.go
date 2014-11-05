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

func (_ *CoreSuite) TestDockerfile_GetWithEtcd(c *C) {
	go startEtcdMockServer()

	d := NewDockerfile(
		[]byte("foo $ETCD_foo qux $FOO $ETCD_bar__foo $ETCD_MISSING $ETCD_foo"),
		&Project{Name: "foo", Repository: "qux"},
		Revision{"foo": "baz"},
		&Environment{Name: "bar", EtcdServers: []string{"http://127.0.0.1:3000"}},
	)

	c.Assert(string(d.Get()), Equals, "foo foofoo qux $FOO barfoobarfoo $ETCD_MISSING foofoo")
}
