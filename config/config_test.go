package config

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type ConfigSuite struct{}

var _ = Suite(&ConfigSuite{})

func (s *ConfigSuite) TestConfig_LoadFile(c *C) {
	var config Config
	err := config.LoadFile("../example/config.ini")

	c.Assert(err, Equals, nil)
	c.Assert(config.Projects, HasLen, 2)

	projectA := config.Projects["project"]
	c.Assert(projectA.GithubToken, Equals, "<your-github-token>")
	c.Assert(projectA.UseShortRevisions, Equals, true)

	projectB := config.Projects["other-project"]
	c.Assert(projectB.GithubToken, Equals, "<other-github-token>")
	c.Assert(projectB.UseShortRevisions, Equals, true)

	c.Assert(projectA.Environments, HasLen, 2)
	c.Assert(projectB.Environments, HasLen, 1)

	c.Assert(
		projectA.Environments["live"].DockerEndPoints[0],
		Equals,
		"http://live-docker.my-company.com:4243",
	)

	c.Assert(
		projectA.Environments["testing"].DockerEndPoints[0],
		Equals,
		"http://testing-docker.my-company.com:4243",
	)

	c.Assert(projectA.Environments["testing"].EtcdServers, HasLen, 1)
	c.Assert(projectB.Environments["live"].EtcdServers, HasLen, 2)

	c.Assert(projectB.Links, HasLen, 2)
	c.Assert(projectB.Links["project"].Alias, Equals, "alias")
	c.Assert(projectB.Links["project"].Project.Name, Equals, projectA.Name)

	c.Assert(projectB.Links["mysql"].Alias, Equals, "db")
	c.Assert(projectB.Links["mysql"].Project, IsNil)
}
