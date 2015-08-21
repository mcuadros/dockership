package core

import (
	"flag"

	. "gopkg.in/check.v1"
)

var githubFlag = flag.Bool("github", false, "Skips Github tests")

func (s *CoreSuite) TestGithub_GetLastRevision(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	p := &Project{
		Repository:          "git@github.com:mcuadros/go-syslog.git",
		RelatedRepositories: []VCS{"git@github.com:mcuadros/go-version.git"},
	}

	g := NewGithub("")
	revision, err := g.GetLastRevision(p)
	c.Assert(err, Equals, nil)
	c.Assert(revision.Get(), Equals, "f7e051ff8c42a6b7cc20b1da6f09de22")
}

func (s *CoreSuite) TestGithub_GetLastCommit(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	p := &Project{
		Repository: "git@github.com:mcuadros/go-syslog.git",
	}

	g := NewGithub("")
	commit, err := g.GetLastCommit(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(commit), Equals, "e079f554382028527e4509d7bb58793b5e98194e")
}

func (s *CoreSuite) TestGithub_GetLastCommitBranch(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	p := &Project{
		Repository: "git@github.com:mcuadros/dockership.git!socket.io",
	}

	g := NewGithub("")
	commit, err := g.GetLastCommit(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(commit), Equals, "1a38193480b3f5fbc10790753f04a406ca460b9c")
}

func (s *CoreSuite) TestGithub_GetDockerFile(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	p := &Project{
		Repository: "git@github.com:mcuadros/dockership.git",
		Dockerfile: ".gitignore",
	}

	g := NewGithub("")
	content, err := g.GetDockerFile(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(content), Equals, "build\nhttp/bindata.go\n")
}

func (s *CoreSuite) TestGithub_GetDockerFileNotFound(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	p := &Project{
		Repository: "git@github.com:github/gem-builder.git",
		Dockerfile: "foo",
	}

	g := NewGithub("")
	_, err := g.GetDockerFile(p)
	c.Assert(err, Not(Equals), nil)
}
