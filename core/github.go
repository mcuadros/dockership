package core

import (
	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
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

func (g *Github) GetDockerFile(p *Project) (content []byte, commit string, err error) {
	commit, err = g.GetLastCommit(p)
	if err != nil {
		return
	}

	content, err = g.getFileContent(p, commit)
	return
}

func (g *Github) GetLastCommit(p *Project) (string, error) {
	c, _, err := g.client.Repositories.GetBranch(p.Owner, p.Repository, p.Branch)
	if err != nil {
		return "", err
	}

	return *c.Commit.SHA, nil
}

func (g *Github) getFileContent(p *Project, commit string) ([]byte, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: commit,
	}

	f, _, _, err := g.client.Repositories.GetContents(p.Owner, p.Repository, p.Dockerfile, opts)
	if err != nil {
		return nil, err
	}

	return f.Decode()
}
