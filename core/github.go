package core

import (
	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"

	. "github.com/mcuadros/dockership/logger"
)

type Github struct {
	client *github.Client
}

func NewGithub(token string) *Github {
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: token},
	}

	return &Github{
		client: github.NewClient(t.Client()),
	}
}

func (g *Github) GetDockerFile(p *Project) (content []byte, commit Commit, err error) {
	commit, err = g.GetLastCommit(p)
	if err != nil {
		return
	}

	content, err = g.getFileContent(p, commit)
	return
}

func (g *Github) GetLastCommit(p *Project) (Commit, error) {
	Debug("Retrieving last commit", "project", p)

	c, r, err := g.client.Repositories.GetBranch(p.Owner, p.Repository, p.Branch)
	if err != nil {
		return "", err
	}

	if r.Remaining < 100 {
		Warning("Low Github request level", "remaining", r.Remaining, "limit", r.Limit)
	}

	return Commit(*c.Commit.SHA), nil
}

func (g *Github) getFileContent(p *Project, commit Commit) ([]byte, error) {
	Debug("Retrieving dockerfile commit", "project", p, "commit", commit)

	opts := &github.RepositoryContentGetOptions{
		Ref: string(commit),
	}

	f, _, r, err := g.client.Repositories.GetContents(p.Owner, p.Repository, p.Dockerfile, opts)
	if err != nil {
		return nil, err
	}

	if r.Remaining < 100 {
		Warning("Low Github request level", "remaining", r.Remaining, "limit", r.Limit)
	}

	return f.Decode()
}

type Commit string

func (c Commit) GetShort() string {
	commit := string(c)
	shortLen := 12
	if len(commit) < shortLen {
		shortLen = len(commit)
	}

	return commit[:shortLen]
}

func (c Commit) String() string {
	return c.GetShort()
}
