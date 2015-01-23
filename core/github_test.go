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
		Repository:          "git@github.com:github/gem-builder.git",
		RelatedRepositories: []VCS{"git@github.com:github/jquery-hotkeys.git"},
	}

	g := NewGithub("")
	revision, err := g.GetLastRevision(p)
	c.Assert(err, Equals, nil)
	c.Assert(revision.Get(), Equals, "c63571780c76de6cf43a04b1ac902f0c")
}

func (s *CoreSuite) TestGithub_GetLastCommit(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	p := &Project{
		Repository: "git@github.com:github/gem-builder.git",
	}

	g := NewGithub("")
	commit, err := g.GetLastCommit(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(commit), Equals, "d170057eca4622d25d3bde81d891ef3f3a2cf060")
}

func (s *CoreSuite) TestGithub_GetLastCommitBranch(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	p := &Project{
		Repository: "git@github.com:github/windows-msysgit.git!diffuse",
	}

	g := NewGithub("")
	commit, err := g.GetLastCommit(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(commit), Equals, "9ef1d6523c8640e04680da27c385d1469e369aa9")
}

func (s *CoreSuite) TestGithub_GetDockerFile(c *C) {
	if !*githubFlag {
		c.Skip("-noGithub not provided")
	}

	p := &Project{
		Repository: "git@github.com:github/gem-builder.git",
		Dockerfile: "git_mock",
	}

	g := NewGithub("")
	content, err := g.GetDockerFile(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(content), Equals, "#!/usr/bin/env ruby\n\n`mkdir -p #{ARGV.last}`\n")
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
