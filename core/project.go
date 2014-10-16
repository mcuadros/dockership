package core

import (
	"fmt"

	. "github.com/mcuadros/dockership/logger"
)

type Project struct {
	GithubToken     string
	Owner           string
	Repository      string
	Branch          string `default:"master"`
	Dockerfile      string `default:"Dockerfile"`
	NoCache         bool
	UseShortCommits bool     `default:"true"`
	EnviromentNames []string `gcfg:"Enviroment"`
	Enviroments     map[string]*Enviroment
}

func (p *Project) Deploy(force bool, enviroment string) error {
	Info("Retrieving last dockerfile ...", "project", p)

	c := NewGithub(p.GithubToken)
	file, commit, err := c.GetDockerFile(p)
	if err != nil {
		Critical(err.Error(), "project", p)
	}

	d := NewDocker(p.mustGetEnviroment(enviroment))
	if err := d.Deploy(p, commit, file, force); err != nil {
		Critical(err.Error(), "project", p, "commit", commit)
		return err
	}

	return nil
}

type ProjectStatus struct {
	Enviroment        *Enviroment
	LastCommit        Commit
	RunningContainers []*Container
	Containers        []*Container
}

func (p *Project) Status() ([]*ProjectStatus, error) {
	r := make([]*ProjectStatus, 0)
	for _, e := range p.Enviroments {
		if s, err := p.StatusByEnviroment(e); err != nil {
			return nil, err
		} else {
			r = append(r, s)
		}
	}

	return r, nil
}

func (p *Project) StatusByEnviroment(enviroment *Enviroment) (*ProjectStatus, error) {
	s := &ProjectStatus{Enviroment: enviroment}

	c := NewGithub(p.GithubToken)
	if commit, err := c.GetLastCommit(p); err != nil {
		return nil, err
	} else {
		s.LastCommit = commit
	}

	d := NewDocker(enviroment)
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
	r := make([]*Container, 0)
	for _, e := range p.Enviroments {
		d := NewDocker(e)
		if l, err := d.ListContainers(p); err != nil {
			return nil, err
		} else {
			r = append(r, l...)
		}
	}

	return r, nil
}

func (p *Project) mustGetEnviroment(name string) *Enviroment {
	if e, ok := p.Enviroments[name]; ok {
		return e
	}

	Critical("Enviroment not defined in project", "project", p, "enviroment", name)
	return nil
}

func (p *Project) String() string {
	return fmt.Sprintf("%s/%s@%s", p.Owner, p.Repository, p.Branch)
}
