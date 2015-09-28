package core

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Github struct {
	client *github.Client
	sync.WaitGroup
}

func NewGithub(token string) *Github {
	var client *http.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)

		client = oauth2.NewClient(oauth2.NoContext, ts)
	}

	return &Github{
		client: github.NewClient(client),
	}
}

func (g *Github) GetDockerFile(p *Project) (content []byte, err error) {
	info := p.Repository.Info()
	commit, err := g.doGetLastCommit(info)
	if err != nil {
		return
	}

	content, err = g.doGetFileContent(info, commit, p.Dockerfile)
	return
}

func (g *Github) GetLastCommit(p *Project) (Commit, error) {
	return g.doGetLastCommit(p.Repository.Info())
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
			commit, err := g.doGetLastCommit(repository.Info())
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

func (g *Github) doGetLastCommit(vcs *VCSInfo) (Commit, error) {
	Debug("Retrieving last commit", "repository", vcs.Origin)
	c, r, err := g.client.Repositories.GetBranch(vcs.Username, vcs.Name, vcs.Branch)
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
