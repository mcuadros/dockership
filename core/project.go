package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mcuadros/go-command"
)

const (
	Deploy Task = "deploy"
)

type Project struct {
	Name                string
	Repository          VCS
	RelatedRepositories []VCS    `gcfg:"RelatedRepository"`
	Dockerfile          string   `default:"Dockerfile"`
	GithubToken         string   `json:"-"`
	History             int      `default:"3"`
	UseShortRevisions   bool     `default:"true"`
	Files               []string `gcfg:"File"`
	TestCommand         string
	NoCache             bool
	Restart             string
	Ports               []string         `gcfg:"Port"`
	Links               map[string]*Link `json:"-"`
	LinkNames           []LinkDefinition `gcfg:"Link"`
	LinkedBy            []*Project       `json:"-"`
	Environments        map[string]*Environment
	EnvironmentNames    []string `gcfg:"Environment"`
	TaskStatus          TaskStatus
	WebHook             string `gcfg:"WebHook"`
}

func (p *Project) Deploy(environment string, output io.Writer, force bool) []error {
	e := p.mustGetEnvironment(environment)
	p.TaskStatus.Start(e, Deploy)
	defer p.TaskStatus.Stop(e, Deploy)

	Info("Retrieving last dockerfile ...", "project", p)

	c := NewGithub(p.GithubToken)
	blob, err := c.GetDockerFile(p)
	if err != nil {
		return []error{err}
	}

	prevStatus, errs := p.StatusByEnvironment(e)
	if len(errs) != 0 {
		return errs
	}
	r, err := c.GetLastRevision(p)
	if err != nil {
		return []error{err}
	}

	d, err := NewDockerGroup(e)
	if err != nil {
		return []error{err}
	}

	file := NewDockerfile(blob, p, r, e)

	errs = d.Deploy(p, r, file, output, force)
	p.afterDeploy(prevStatus, e, errs)
	return errs
}

func (p *Project) afterDeploy(prevStatus *ProjectStatus, e *Environment, errs []error) {
	if p.WebHook == "" {
		return
	}
	go func() {
		prevRev := getRunningRevFromStatus(prevStatus)
		currStatus, _ := p.StatusByEnvironment(e)
		currRev := getRunningRevFromStatus(currStatus)
		errStrings := make([]string, 0, len(errs))
		for _, err := range errs {
			errStrings = append(errStrings, err.Error())
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"project":           p.Name,
			"repository":        p.Repository,
			"environment":       e.Name,
			"previous_revision": prevRev,
			"current_revision":  currRev,
			"errors":            errs,
		})
		Info("Calling WebHook at "+p.WebHook, "project", p)
		http.Post(p.WebHook, "application/json", bytes.NewReader(payload))
	}()
}

func getRunningRevFromStatus(status *ProjectStatus) *string {
	var s *string
	if status != nil && len(status.RunningContainers) > 0 {
		s = &strings.Split(string(status.RunningContainers[0].Image), ":")[1]
	}
	return s
}

func (p *Project) mustGetEnvironment(name string) *Environment {
	if e, ok := p.Environments[name]; ok {
		return e
	}

	Critical("Environment not defined in project", "project", p, "environment", name)
	return nil
}

type ProjectDeployResult struct {
	*command.ExecutionResponse
}

func (p *Project) Test(environment string) (*ProjectDeployResult, error) {
	if p.TestCommand == "" {
		return nil, nil
	}

	Info("Executing Test command", "project", p, "script", p.TestCommand)
	json, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	cmd := command.NewCommand(fmt.Sprintf("%s %s %s", p.TestCommand, environment, json))
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
	Environment       *Environment
	LastRevision      Revision
	RunningContainers []*Container
	Containers        []*Container
}

func (p *Project) Status() ([]*ProjectStatus, []error) {
	e := make([]error, 0)
	r := make([]*ProjectStatus, 0)
	for _, env := range p.Environments {
		if s, err := p.StatusByEnvironment(env); err != nil {
			e = append(e, err...)
		} else {
			r = append(r, s)
		}
	}

	return r, e
}

func (p *Project) StatusByEnvironment(e *Environment) (*ProjectStatus, []error) {
	s := &ProjectStatus{Environment: e}

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

func (p *Project) ListContainers() ([]*Container, []error) {
	e := make([]error, 0)
	r := make([]*Container, 0)
	for _, env := range p.Environments {
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

func (p *Project) ListImages() ([]*Image, []error) {
	e := make([]error, 0)
	r := make([]*Image, 0)
	for _, env := range p.Environments {
		d, err := NewDockerGroup(env)
		if err != nil {
			e = append(e, err)
		} else {
			if l, err := d.ListImages(p); len(err) != 0 {
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
	return fmt.Sprintf("%s/%s!%s", i.Username, i.Name, i.Branch)
}
