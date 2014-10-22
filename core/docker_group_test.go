package core

import (
	"github.com/fsouza/go-dockerclient/testing"
	. "gopkg.in/check.v1"
)

func (s *CoreSuite) TestDockerGroup_BuildImage(c *C) {
	dg := &DockerGroup{dockers: make(map[string]*Docker, 0)}
	for i := 0; i < 5; i++ {
		m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
		dg.dockers[m.URL()], _ = NewDocker(m.URL())
		m.Stop()
	}

	p := &Project{Repository: "git@github.com:foo/bar.git", UseShortRevisions: true}
	r := Revision{"foo/bar": Commit("qux")}
	result := dg.BuildImage(p, r, []byte(""))

	c.Assert(result, HasLen, 5)
	for _, r := range result {
		c.Assert(r, ErrorMatches, "cannot connect to Docker endpoint")
	}
}

func (s *CoreSuite) TestDockerGroup_Run(c *C) {
	dg := &DockerGroup{dockers: make(map[string]*Docker, 0)}
	for i := 0; i < 5; i++ {
		m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
		dg.dockers[m.URL()], _ = NewDocker(m.URL())
		m.Stop()
	}

	p := &Project{Repository: "git@github.com:foo/bar.git", UseShortRevisions: true}
	r := Revision{"foo/bar": Commit("qux")}
	result := dg.Run(p, r)

	c.Assert(result, HasLen, 5)
	for _, r := range result {
		c.Assert(r, ErrorMatches, "cannot connect to Docker endpoint")
	}
}

func (s *CoreSuite) TestDockerGroup_Clean(c *C) {
	dg := &DockerGroup{dockers: make(map[string]*Docker, 0)}
	for i := 0; i < 5; i++ {
		m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
		dg.dockers[m.URL()], _ = NewDocker(m.URL())
		m.Stop()
	}

	p := &Project{Repository: "git@github.com:foo/bar.git", UseShortRevisions: true}
	result := dg.Clean(p)

	c.Assert(result, HasLen, 5)
	for _, r := range result {
		c.Assert(r, ErrorMatches, "cannot connect to Docker endpoint")
	}
}

func (s *CoreSuite) TestDockerGroup_DeployAndListContainers(c *C) {
	dg := &DockerGroup{dockers: make(map[string]*Docker, 0)}
	for i := 0; i < 5; i++ {
		m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
		defer m.Stop()
		dg.dockers[m.URL()], _ = NewDocker(m.URL())
	}

	p := &Project{Repository: "git@github.com:foo/bar.git", UseShortRevisions: true}
	r := Revision{"foo/bar": Commit("qux")}

	errors := dg.Deploy(p, r, []byte(""), true)
	c.Assert(errors, HasLen, 0)

	result, errors := dg.ListContainers(p)
	c.Assert(errors, HasLen, 0)

	c.Assert(result, HasLen, 5)
	for _, r := range result {
		c.Assert(r.Image, Equals, ImageId("foo/bar:qux"))
	}
}
