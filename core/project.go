package core

import (
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/mcuadros/dockership/logger"

	"github.com/mcuadros/go-command"
)

type Project struct {
	Name                string
	Repository          VCS
	RelatedRepositories []VCS  `gcfg:"RelatedRepository"`
	Branch              string `default:"master"`
	Dockerfile          string `default:"Dockerfile"`
	NoCache             bool
	UseShortRevisions   bool     `default:"true"`
	Ports               []string `gcfg:"Port"`
	Files               []string `gcfg:"File"`
	Enviroments         map[string]*Enviroment
	EnviromentNames     []string `gcfg:"Enviroment"`
	TestCommand         string
	GithubToken         string
}

func (p *Project) Deploy(enviroment string, force bool) (*ProjectDeployResult, error) {
	Info("Retrieving last dockerfile ...", "project", p)

	c := NewGithub(p.GithubToken)
	file, err := c.GetDockerFile(p)
	if err != nil {
		return nil, err
	}

	rev, err := c.GetLastRevision(p)
	if err != nil {
		return nil, err
	}

	d := NewDocker(p.mustGetEnviroment(enviroment))
	if err := d.Deploy(p, rev, file, force); err != nil {
		return nil, err
	}

	return p.Test(enviroment)
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
	if rev, err := c.GetLastRevision(p); err != nil {
		return nil, err
	} else {
		s.LastRevision = rev
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

func (p *Project) String() string {
	i := p.Repository.Info()
	return fmt.Sprintf("%s/%s@%s", i.Username, i.Name, p.Branch)
}
