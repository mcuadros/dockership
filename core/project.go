package core

import (
	"fmt"

	. "github.com/mcuadros/dockership/logger"
)

type Project struct {
	GithubToken     string
	DockerEndPoint  string
	Owner           string
	Repository      string
	Branch          string `default:"master"`
	Dockerfile      string `default:"Dockerfile"`
	NoCache         bool
	UseShortCommits bool `default:"true"`
}

func (p *Project) Deploy(force bool) error {
	Info("Retrieving last dockerfile ...", "project", p)

	c := NewGithub(p.GithubToken)
	file, commit, err := c.GetDockerFile(p)
	if err != nil {
		Critical(err.Error(), "project", p)
	}

	d := NewDocker(p.DockerEndPoint)
	if err := d.Deploy(p, commit, file, force); err != nil {
		Critical(err.Error(), "project", p, "commit", commit)
		return err
	}

	return nil
}

type ProjectStatus struct {
	LastCommit        Commit
	RunningContainers []*Container
	Containers        []*Container
}

func (p *Project) Status() (*ProjectStatus, error) {
	s := &ProjectStatus{}

	c := NewGithub(p.GithubToken)
	if commit, err := c.GetLastCommit(p); err != nil {
		return nil, err
	} else {
		s.LastCommit = commit
	}

	d := NewDocker(p.DockerEndPoint)
	if l, err := d.ListContainers(p); err != nil {
		return nil, err
	} else {
		s.Containers = l
	}

	s.RunningContainers = make([]*Container, 0)
	for _, c := range s.Containers {
		if c.IsRunning() {
			s.RunningContainers = append(s.RunningContainers, c)
		}
	}

	return s, nil
}

func (p *Project) List() ([]*Container, error) {
	d := NewDocker(p.DockerEndPoint)
	return d.ListContainers(p)
}

func (p *Project) String() string {
	return fmt.Sprintf("%s/%s@%s", p.Owner, p.Repository, p.Branch)
}
