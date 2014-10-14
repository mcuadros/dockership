package main

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

func (g *Github) GetDockerFile(owner, repository, branch, dockerfile string) (content []byte, commit string, err error) {
	commit, err = g.getLastCommit(branch)
	if err != nil {
		return
	}

	content, err = g.getFileContent(owner, repository, dockerfile, commit)
	return
}

func (g *Github) getFileContent(owner, repository, dockerfile, commit string) ([]byte, error) {
	opts := &github.RepositoryContentGetOptions{
		Ref: commit,
	}

	f, _, _, err := g.client.Repositories.GetContents(owner, repository, dockerfile, opts)
	if err != nil {
		return nil, err
	}

	return f.Decode()
}

func (g *Github) getLastCommit(branch string) (string, error) {
	c, _, err := g.client.Repositories.GetBranch(owner, repository, branch)
	if err != nil {
		return "", err
	}

	return *c.Commit.SHA, nil
}
