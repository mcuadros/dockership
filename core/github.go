package core

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"sync"

	"github.com/google/go-github/github"

	. "github.com/mcuadros/dockership/logger"
)

type Github struct {
	client *github.Client
	sync.WaitGroup
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
	Verbose()
	Debug("Retrieving last commit", "project", p)

	c := make(chan string)
	h := func(owner, repository, branch string) {
		commit, err := g.doGetLastCommit(owner, repository, branch)
		if err != nil {
			panic(err)
		}
		c <- commit
		g.Done()
	}

	g.Add(1)
	go h(p.Owner, p.Repository, p.Branch)
	g.Wait()

	stack := make([]string, 0)
	for commit := range c {
		fmt.Println(commit)
		stack = append(stack, commit)
	}

	fmt.Println(stack)

	return Commit(stack[0]), nil
}

func (g *Github) doGetLastCommit(owner, repository, branch string) (string, error) {

	Debug("Retrieving last commit", "project", owner, "repository", repository, "branch", branch)

	c, r, err := g.client.Repositories.GetBranch(owner, repository, branch)
	if err != nil {
		return "", err
	}

	if r.Remaining < 100 {
		Warning("Low Github request level", "remaining", r.Remaining, "limit", r.Limit)
	}

	return string(*c.Commit.SHA), nil
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
