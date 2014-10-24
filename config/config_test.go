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
	c.Assert(config.Projects, HasLen, 1)

	project := config.Projects["project"]
	c.Assert(project.GithubToken, Equals, "<your-github-token>")
	c.Assert(project.UseShortRevisions, Equals, true)

	c.Assert(project.Enviroments, HasLen, 2)

	c.Assert(
		project.Enviroments["live"].DockerEndPoints[0],
		Equals,
		"http://live-docker.my-company.com:4243",
	)

	c.Assert(
		project.Enviroments["testing"].DockerEndPoints[0],
		Equals,
		"http://testing-docker.my-company.com:4243",
	)
}
