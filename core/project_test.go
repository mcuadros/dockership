package core

import (
	"bytes"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient/testing"
	. "gopkg.in/check.v1"
)

func (s *CoreSuite) TestProject_Deploy(c *C) {
	if !*githubFlag {
		c.Skip("--github not provided")
	}

	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	e := &Environment{Name: "a", DockerEndPoints: []string{m.URL()}}
	p := &Project{
		Repository:   "git@github.com:mcuadros/dockership.git",
		Environments: map[string]*Environment{"foo": e},
		Dockerfile:   "Dockerfile",
		GithubToken:  githubToken,
		TaskStatus:   TaskStatus{},
	}

	input := bytes.NewBuffer(nil)
	err := p.Deploy("foo", input, false)
	c.Assert(err, HasLen, 0)

	l, err := p.ListContainers()
	c.Assert(err, HasLen, 0)
	c.Assert(l, HasLen, 1)
	c.Assert(l[0].DockerEndPoint, Equals, e.DockerEndPoints[0])
	c.Assert(string(input.Bytes()), HasLen, 51)
}

func (s *CoreSuite) TestProject_Test(c *C) {
	p := &Project{
		Repository:   "git@github.com:foo/bar.git",
		Environments: map[string]*Environment{"a": &Environment{}},
		TestCommand:  "foo",
	}

	_, err := p.Test("a")
	c.Assert(err, Not(Equals), nil)
}

func (s *CoreSuite) TestProject_TestFail(c *C) {
	p := &Project{
		Repository:   "git@github.com:foo/bar.git",
		Environments: map[string]*Environment{"a": &Environment{}},
		TestCommand:  "echo",
	}

	r, err := p.Test("a")
	c.Assert(err, Equals, nil)
	c.Assert(strings.HasPrefix(string(r.Stdout), "a"), Equals, true)
}

func (s *CoreSuite) TestProject_Status(c *C) {
	if !*githubFlag {
		c.Skip("--github not provided")
	}

	mA, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	mB, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	envs := map[string]*Environment{
		"a": &Environment{Name: "a", DockerEndPoints: []string{mA.URL()}},
		"b": &Environment{Name: "b", DockerEndPoints: []string{mB.URL()}},
	}

	p := &Project{
		Name:         "foo",
		Repository:   "git@github.com:mcuadros/cli-array-editor.git",
		GithubToken:  githubToken,
		Environments: envs,
		Dockerfile:   ".gitignore",
	}

	input := bytes.NewBuffer(nil)
	da, _ := NewDocker(envs["a"].DockerEndPoints[0], nil)
	da.Deploy(p, Revision{}, &Dockerfile{}, input, false)
	db, _ := NewDocker(envs["b"].DockerEndPoints[0], nil)
	db.Deploy(p, Revision{}, &Dockerfile{}, input, false)

	r, err := p.Status()
	c.Assert(err, HasLen, 0)
	c.Assert(r, HasLen, 2)
	c.Assert(r[0].Environment, Equals, envs["a"])
	c.Assert(r[0].LastRevision.GetShort(), Equals, "a44ffbb10515")
	c.Assert(r[0].Containers, HasLen, 1)
	c.Assert(r[0].RunningContainers, HasLen, 1)
	c.Assert(r[0].Containers[0], Equals, r[0].RunningContainers[0])
}

func (s *CoreSuite) TestProject_ListContainers(c *C) {
	mA, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	mB, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	envs := map[string]*Environment{
		"a": &Environment{Name: "a", DockerEndPoints: []string{mA.URL()}},
		"b": &Environment{Name: "b", DockerEndPoints: []string{mB.URL()}},
	}

	p := &Project{
		Name:         "foo",
		Repository:   "git@github.com:foo/bar.git",
		Environments: envs,
	}

	input := bytes.NewBuffer(nil)

	da, _ := NewDocker(envs["a"].DockerEndPoints[0], nil)
	da.Deploy(p, Revision{}, &Dockerfile{}, input, false)
	db, _ := NewDocker(envs["b"].DockerEndPoints[0], nil)
	db.Deploy(p, Revision{}, &Dockerfile{}, input, false)
	time.Sleep(1 * time.Second)
	l, err := p.ListContainers()
	c.Assert(err, HasLen, 0)
	c.Assert(l, HasLen, 2)
	c.Assert(l[0].APIContainers.Names[0], Equals, "/foo")
	c.Assert(l[1].APIContainers.Names[0], Equals, "/foo")
}

func (s *CoreSuite) TestProject_ListImages(c *C) {
	mA, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	mB, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	envs := map[string]*Environment{
		"a": &Environment{Name: "a", DockerEndPoints: []string{mA.URL()}},
		"b": &Environment{Name: "b", DockerEndPoints: []string{mB.URL()}},
	}

	p := &Project{
		Repository:   "git@github.com:foo/bar.git",
		Environments: envs,
	}

	input := bytes.NewBuffer(nil)

	da, _ := NewDocker(envs["a"].DockerEndPoints[0], nil)
	da.Deploy(p, Revision{}, &Dockerfile{}, input, false)
	db, _ := NewDocker(envs["b"].DockerEndPoints[0], nil)
	db.Deploy(p, Revision{}, &Dockerfile{}, input, false)
	time.Sleep(1 * time.Second)
	l, err := p.ListImages()
	c.Assert(err, HasLen, 0)
	c.Assert(l, HasLen, 2)
	c.Assert(l[0].APIImages.ID, Not(Equals), "")
	c.Assert(l[1].APIImages.ID, Not(Equals), "")
}
