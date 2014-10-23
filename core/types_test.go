package core

import (
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/fsouza/go-dockerclient"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type CoreSuite struct {
	sync.WaitGroup
}

var _ = Suite(&CoreSuite{})

func (s *CoreSuite) TestVCS_IsValid(c *C) {
	vcs := VCS("git@github.com:mcuadros/dockership.git")
	c.Assert(vcs.IsValid(), Equals, true)
}

func (s *CoreSuite) TestVCS_NotIsValid(c *C) {
	vcs := VCS("foo")
	c.Assert(vcs.IsValid(), Equals, false)
}

func (s *CoreSuite) TestVCS_Info(c *C) {
	vcs := VCS("git@github.com:mcuadros/dockership.git")

	c.Assert(vcs.Info().Name, Equals, "dockership")
	c.Assert(vcs.Info().Username, Equals, "mcuadros")
	c.Assert(vcs.Info().Branch, Equals, "master")
}

func (s *CoreSuite) TestVCS_InfoBranch(c *C) {
	vcs := VCS("git@github.com:mcuadros/dockership.git!branch")

	c.Assert(vcs.Info().Name, Equals, "dockership")
	c.Assert(vcs.Info().Username, Equals, "mcuadros")
	c.Assert(vcs.Info().Branch, Equals, "branch")
}

func (s *CoreSuite) TestRevision_GetShort(c *C) {
	revision := Revision{"foo": "123456789123456789", "bar:": "123456789123456789"}

	c.Assert(revision.GetShort(), Equals, "28a247e8ba3a")
}

func (s *CoreSuite) TestRevision_Get(c *C) {
	revision := Revision{"foo": "123456789123456789", "bar:": "123456789123456789"}

	c.Assert(revision.Get(), Equals, "28a247e8ba3ab48ab72dd196f1052f8a")
}

func (s *CoreSuite) TestRevision_GetOneKey(c *C) {
	revision := Revision{"foo": "123456789123456789"}

	c.Assert(revision.Get(), Equals, "123456789123456789")
}

func (s *CoreSuite) TestRevision_String(c *C) {
	revisionAZ := Revision{"foo": "ZZZZZZZZZZZZZZZZZZ", "bar:": "123456789123456789"}
	c.Assert(fmt.Sprintf("%s", revisionAZ), Equals, "e1ba1f05de5f184fe94ec745250b5d9e")

	revisionZA := Revision{"foo": "123456789123456789", "bar:": "ZZZZZZZZZZZZZZZZZZ"}
	c.Assert(fmt.Sprintf("%s", revisionZA), Equals, "e1ba1f05de5f184fe94ec745250b5d9e")
}

func (s *CoreSuite) TestImageId_IsCommit(c *C) {
	i := ImageId("foo/bar:bar")

	c.Assert(i.IsRevision(Revision{"foo": "bar"}), Equals, true)
	c.Assert(i.IsRevision(Revision{"foo": "qux"}), Equals, false)
}

func (s *CoreSuite) TestImageId_BelongsTo(c *C) {
	i := ImageId("foo/bar:qux")

	c.Assert(i.BelongsTo(&Project{
		Repository: "git@github.com:foo/bar.git",
	}), Equals, true)

	c.Assert(i.BelongsTo(&Project{
		Repository: "git@github.com:qux/bar.git",
	}), Equals, false)
}

func (s *CoreSuite) TestImageId_GetRevisionString(c *C) {
	i := ImageId("foo/bar:qux")

	c.Assert(i.GetRevisionString(), Equals, "qux")
}

func (s *CoreSuite) TestImage_BelongsTo(c *C) {
	i := Image{
		APIImages: docker.APIImages{
			RepoTags: []string{"foo/bar:qux"},
		},
	}

	c.Assert(i.BelongsTo(&Project{
		Repository: "git@github.com:foo/bar.git",
	}), Equals, true)

	c.Assert(i.BelongsTo(&Project{
		Repository: "git@github.com:qux/bar.git",
	}), Equals, false)
}

func (s *CoreSuite) TestImageId_GetProjectString(c *C) {
	i := ImageId("foo/bar:qux")

	c.Assert(i.GetProjectString(), Equals, "foo/bar")
}

func (s *CoreSuite) TestContainer_IsRunningUp(c *C) {
	co := Container{APIContainers: docker.APIContainers{Status: "Up foo"}}

	c.Assert(co.IsRunning(), Equals, true)
}

func (s *CoreSuite) TestContainer_IsRunningDown(c *C) {
	co := Container{APIContainers: docker.APIContainers{Status: "foo"}}

	c.Assert(co.IsRunning(), Equals, false)
}

func (s *CoreSuite) TestContainer_GetShortIdLong(c *C) {
	co := Container{APIContainers: docker.APIContainers{ID: "123456789123456789"}}

	c.Assert(co.GetShortId(), Equals, "123456789123")
}

func (s *CoreSuite) TestContainer_GetShortId(c *C) {
	co := Container{APIContainers: docker.APIContainers{ID: "123456"}}

	c.Assert(co.GetShortId(), Equals, "123456")
}

func (s *CoreSuite) TestContainer_GetPortsString(c *C) {
	co := Container{APIContainers: docker.APIContainers{
		Ports: []docker.APIPort{
			docker.APIPort{PrivatePort: 42, PublicPort: 84, Type: "tcp", IP: "0.0.0.0"},
			docker.APIPort{PrivatePort: 42, PublicPort: 84, Type: "tcp"},
		},
	}}

	c.Assert(co.GetPortsString(), Equals, "0.0.0.0:84->42/tcp, 42/tcp")
}

func (s *CoreSuite) TestContainersByCreated_Sort(c *C) {
	list := []*Container{
		&Container{APIContainers: docker.APIContainers{Created: 3}},
		&Container{APIContainers: docker.APIContainers{Created: 1}},
		&Container{APIContainers: docker.APIContainers{Created: 2}},
	}

	sort.Sort(ContainersByCreated(list))

	c.Assert(list[0].Created, Equals, int64(1))
	c.Assert(list[2].Created, Equals, int64(3))
}

func (s *CoreSuite) TestEnviroment_String(c *C) {
	e := Enviroment{Name: "foo"}

	c.Assert(e.String(), Equals, "foo")
}
