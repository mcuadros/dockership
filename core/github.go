package core

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

type Github struct {
	client *github.Client
	sync.WaitGroup
}

func NewGithub(token string) *Github {
	var client *http.Client

	if token != "" {
		t := &oauth.Transport{
			Token: &oauth.Token{AccessToken: token},
		}

		client = t.Client()
	}

	return &Github{
		client: github.NewClient(client),
	}
}

func (g *Github) GetDockerFile(p *Project) (content []byte, err error) {
	info := p.Repository.Info()
	commit, err := g.doGetLastCommit(info, p.Branch)
	if err != nil {
		return
	}

	content, err = g.doGetFileContent(info, commit, p.Dockerfile)
	return
}

func (g *Github) GetLastCommit(p *Project) (Commit, error) {
	return g.doGetLastCommit(p.Repository.Info(), p.Branch)
}

func (g *Github) GetLastRevision(p *Project) (Revision, error) {
	repos := p.RelatedRepositories
	repos = append(repos, p.Repository)
	count := len(repos)

	type msg struct {
		repository VCS
		commit     Commit
		err        error
	}

	c := make(chan msg, count)
	defer close(c)

	for _, repository := range repos {
		g.Add(1)
		go func(repository VCS) {
			defer g.Done()
			commit, err := g.doGetLastCommit(repository.Info(), p.Branch)
			c <- msg{repository, commit, err}
		}(repository)
	}

	g.Wait()

	revision := make(Revision, 0)
	for i := 0; i < count; i++ {
		m := <-c
		if m.err != nil {
			return nil, m.err
		}

		revision[m.repository] = m.commit
	}

	return revision, nil
}

func (g *Github) doGetLastCommit(vcs *VCSInfo, branch string) (Commit, error) {
	Debug("Retrieving last commit", "repository", vcs.Origin)
	c, r, err := g.client.Repositories.GetBranch(vcs.Username, vcs.Name, branch)
	if err != nil {
		return "", err
	}

	if r.Remaining < 100 {
		Warning("Low Github request level", "remaining", r.Remaining, "limit", r.Limit)
	}

	return Commit(*c.Commit.SHA), nil
}

func (g *Github) doGetFileContent(vcs *VCSInfo, commit Commit, file string) ([]byte, error) {
	Debug("Retrieving dockerfile commit", "repository", vcs.Origin, "commit", commit)
	opts := &github.RepositoryContentGetOptions{
		Ref: string(commit),
	}

	f, _, r, err := g.client.Repositories.GetContents(vcs.Username, vcs.Name, file, opts)
	if err != nil {
		return nil, err
	}

	if r.Remaining < 100 {
		Warning("Low Github request level", "remaining", r.Remaining, "limit", r.Limit)
	}

	if f == nil {
		return nil, errors.New(fmt.Sprintf("Unable to find %q file", file))
	}

	return f.Decode()
}
