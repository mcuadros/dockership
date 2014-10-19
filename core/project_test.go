package core

import (
	"strings"

	"github.com/fsouza/go-dockerclient/testing"
	. "gopkg.in/check.v1"
)

func (s *CoreSuite) TestProject_Deploy(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	e := &Enviroment{Name: "a", DockerEndPoint: m.URL()}
	p := &Project{
		Repository:  "git@github.com:github/gem-builder.git",
		Branch:      DEFAULT_BRANCH,
		Enviroments: map[string]*Enviroment{"foo": e},
		Dockerfile:  "git_mock",
		GithubToken: "05bed21c257d935017d85d3398b46ac81035756f",
	}

	_, err := p.Deploy("foo", false)
	c.Assert(err, Equals, nil)

	l, err := p.List()
	c.Assert(err, Equals, nil)
	c.Assert(l, HasLen, 1)
	c.Assert(l[0].Enviroment.DockerEndPoint, Equals, e.DockerEndPoint)
}

func (s *CoreSuite) TestProject_Test(c *C) {
	p := &Project{
		Repository:  "git@github.com:foo/bar.git",
		Branch:      DEFAULT_BRANCH,
		Enviroments: map[string]*Enviroment{"a": &Enviroment{}},
		TestCommand: "foo",
	}

	_, err := p.Test("a")
	c.Assert(err, Not(Equals), nil)
}

func (s *CoreSuite) TestProject_TestFail(c *C) {
	p := &Project{
		Repository:  "git@github.com:foo/bar.git",
		Branch:      DEFAULT_BRANCH,
		Enviroments: map[string]*Enviroment{"a": &Enviroment{}},
		TestCommand: "echo",
	}

	r, err := p.Test("a")
	c.Assert(err, Equals, nil)
	c.Assert(strings.HasPrefix(string(r.Stdout), "a"), Equals, true)
}

func (s *CoreSuite) TestProject_Status(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	mA, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	mB, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	envs := map[string]*Enviroment{
		"a": &Enviroment{Name: "a", DockerEndPoint: mA.URL()},
		"b": &Enviroment{Name: "b", DockerEndPoint: mB.URL()},
	}

	p := &Project{
		Repository:  "git@github.com:github/gem-builder.git",
		Branch:      DEFAULT_BRANCH,
		Enviroments: envs,
		Dockerfile:  "git_mock",
	}

	NewDocker(envs["a"]).Deploy(p, Revision{}, []byte{}, false)
	NewDocker(envs["b"]).Deploy(p, Revision{}, []byte{}, false)

	r, err := p.Status()
	c.Assert(err, Equals, nil)
	c.Assert(r, HasLen, 2)
	c.Assert(r[0].Enviroment, Equals, envs["a"])
	c.Assert(r[0].LastRevision.GetShort(), Equals, "d170057eca46")
	c.Assert(r[0].Containers, HasLen, 1)
	c.Assert(r[0].RunningContainers, HasLen, 1)
	c.Assert(r[0].Containers[0], Equals, r[0].RunningContainers[0])
}

func (s *CoreSuite) TestProject_List(c *C) {
	mA, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	mB, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	envs := map[string]*Enviroment{
		"a": &Enviroment{Name: "a", DockerEndPoint: mA.URL()},
		"b": &Enviroment{Name: "b", DockerEndPoint: mB.URL()},
	}

	p := &Project{
		Repository:  "git@github.com:foo/bar.git",
		Branch:      DEFAULT_BRANCH,
		Enviroments: envs,
	}

	NewDocker(envs["a"]).Deploy(p, Revision{}, []byte{}, false)
	NewDocker(envs["b"]).Deploy(p, Revision{}, []byte{}, false)

	l, err := p.List()
	c.Assert(err, Equals, nil)
	c.Assert(l, HasLen, 2)
	c.Assert(l[0].Enviroment.DockerEndPoint, Equals, envs["a"].DockerEndPoint)
	c.Assert(l[1].Enviroment.DockerEndPoint, Equals, envs["b"].DockerEndPoint)

}