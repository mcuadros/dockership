package core

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mcuadros/go-command"
)

type Project struct {
	Name                string
	Repository          VCS
	RelatedRepositories []VCS    `gcfg:"RelatedRepository"`
	Branch              string   `default:"master"`
	Dockerfile          string   `default:"Dockerfile"`
	GithubToken         string   `json:"-"`
	History             int      `default:"3"`
	UseShortRevisions   bool     `default:"true"`
	Files               []string `gcfg:"File"`
	TestCommand         string
	NoCache             bool
	Ports               []string               `gcfg:"Port"`
	Links               map[string]*Link       `json:"-"`
	LinkNames           []LinkDefinition       `gcfg:"Link"`
	LinkedBy            []*Project             `json:"-"`
	Enviroments         map[string]*Enviroment `json:"-"`
	EnviromentNames     []string               `gcfg:"Enviroment"`
}

func (p *Project) Deploy(enviroment string, force bool) []error {
	Info("Retrieving last dockerfile ...", "project", p)

	c := NewGithub(p.GithubToken)
	file, err := c.GetDockerFile(p)
	if err != nil {
		return []error{err}
	}

	rev, err := c.GetLastRevision(p)
	if err != nil {
		return []error{err}
	}

	d, err := NewDockerGroup(p.mustGetEnviroment(enviroment))
	if err != nil {
		return []error{err}
	}

	return d.Deploy(p, rev, file, force)
}

func (p *Project) mustGetEnviroment(name string) *Enviroment {
	if e, ok := p.Enviroments[name]; ok {
		return e
	}

	Critical("Enviroment not defined in project", "project", p, "enviroment", name)
	return nil
}

type ProjectDeployResult struct {
	*command.ExecutionResponse
}

func (p *Project) Test(enviroment string) (*ProjectDeployResult, error) {
	if p.TestCommand == "" {
		return nil, nil
	}

	Info("Executing Test command", "project", p, "script", p.TestCommand)
	json, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	cmd := command.NewCommand(fmt.Sprintf("%s %s %s", p.TestCommand, enviroment, json))
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	response := cmd.GetResponse()
	result := &ProjectDeployResult{ExecutionResponse: response}

	if response.Failed {
		return result, errors.New(fmt.Sprintf("Test script %q failed", p.TestCommand))
	}

	return result, nil
}

type ProjectStatus struct {
	Enviroment        *Enviroment
	LastRevision      Revision
	RunningContainers []*Container
	Containers        []*Container
}

func (p *Project) Status() ([]*ProjectStatus, []error) {
	e := make([]error, 0)
	r := make([]*ProjectStatus, 0)
	for _, env := range p.Enviroments {
		if s, err := p.StatusByEnviroment(env); err != nil {
			e = append(e, err...)
		} else {
			r = append(r, s)
		}
	}

	return r, e
}

func (p *Project) StatusByEnviroment(e *Enviroment) (*ProjectStatus, []error) {
	s := &ProjectStatus{Enviroment: e}

	c := NewGithub(p.GithubToken)
	if rev, err := c.GetLastRevision(p); err != nil {
		return nil, []error{err}
	} else {
		s.LastRevision = rev
	}

	d, err := NewDockerGroup(e)
	if err != nil {
		return nil, []error{err}
	}

	if l, err := d.ListContainers(p); len(err) != 0 {
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

func (p *Project) List() ([]*Container, []error) {
	e := make([]error, 0)
	r := make([]*Container, 0)
	for _, env := range p.Enviroments {
		d, err := NewDockerGroup(env)
		if err != nil {
			e = append(e, err)
		} else {
			if l, err := d.ListContainers(p); len(err) != 0 {
				e = append(e, err...)
			} else {
				r = append(r, l...)
			}
		}
	}

	return r, e
}

func (p *Project) String() string {
	i := p.Repository.Info()
	return fmt.Sprintf("%s/%s@%s", i.Username, i.Name, p.Branch)
}
