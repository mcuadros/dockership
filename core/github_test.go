package core

import (
	"flag"
	"os"

	. "gopkg.in/check.v1"
)

var githubToken = os.Getenv("GITHUB_API_TOKEN")
var githubFlag = flag.Bool("github", false, "Skips Github tests")

func (s *CoreSuite) TestGithub_GetLastRevision(c *C) {
	if !*githubFlag {
		c.Skip("-github not provided")
	}

	p := &Project{
		Repository:          "git@github.com:mcuadros/cli-array-editor.git",
		RelatedRepositories: []VCS{"git@github.com:mcuadros/silex-hateoas.git"},
	}

	g := NewGithub(githubToken)
	revision, err := g.GetLastRevision(p)
	c.Assert(err, Equals, nil)
	c.Assert(revision.Get(), Equals, "476e1056780a5912677ec4864478601d")
}

func (s *CoreSuite) TestGithub_GetLastCommit(c *C) {
	if !*githubFlag {
		c.Skip("-github not provided")
	}

	p := &Project{
		Repository: "git@github.com:mcuadros/cli-array-editor.git",
	}

	g := NewGithub(githubToken)
	commit, err := g.GetLastCommit(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(commit), Equals, "a44ffbb10515ea575056703238114463034131ca")
}

func (s *CoreSuite) TestGithub_GetLastCommitBranch(c *C) {
	if !*githubFlag {
		c.Skip("-github not provided")
	}

	p := &Project{
		Repository: "git@github.com:mcuadros/dockership.git!socket.io",
	}

	g := NewGithub(githubToken)
	commit, err := g.GetLastCommit(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(commit), Equals, "1a38193480b3f5fbc10790753f04a406ca460b9c")
}

func (s *CoreSuite) TestGithub_GetDockerFile(c *C) {
	if !*githubFlag {
		c.Skip("-github not provided")
	}

	p := &Project{
		Repository: "git@github.com:mcuadros/dockership.git",
		Dockerfile: ".gitignore",
	}

	g := NewGithub(githubToken)
	content, err := g.GetDockerFile(p)
	c.Assert(err, Equals, nil)
	c.Assert(string(content), Equals, "build\nhttp/bindata.go\n")
}

func (s *CoreSuite) TestGithub_GetDockerFileNotFound(c *C) {
	if !*githubFlag {
		c.Skip("-github not provided")
	}

	p := &Project{
		Repository: "git@github.com:github/gem-builder.git",
		Dockerfile: "foo",
	}

	g := NewGithub(githubToken)
	_, err := g.GetDockerFile(p)
	c.Assert(err, Not(Equals), nil)
}
