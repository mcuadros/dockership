package core

import (
	"encoding/json"
	"fmt"

	. "github.com/mcuadros/dockership/logger"

	"github.com/mcuadros/go-command"
)

type Project struct {
	Name                string
	Repository          VCS
	RelatedRepositories []VCS `gcfg:"RelatedRepository"`

	Owner           string
	GithubToken     string
	Branch          string `default:"master"`
	Dockerfile      string `default:"Dockerfile"`
	NoCache         bool
	Ports           []string `gcfg:"Port"`
	UseShortCommits bool     `default:"true"`
	EnviromentNames []string `gcfg:"Enviroment"`
	Enviroments     map[string]*Enviroment
	TestCommand     string
	Files           []string `gcfg:"File"`
}

func (p *Project) Deploy(force bool, enviroment string) error {
	Info("Retrieving last dockerfile ...", "project", p)

	c := NewGithub(p.GithubToken)
	file, err := c.GetDockerFile(p)
	if err != nil {
		Critical(err.Error(), "project", p)
	}

	rev, err := c.GetLastRevision(p)
	if err != nil {
		Critical(err.Error(), "project", p)
	}

	d := NewDocker(p.mustGetEnviroment(enviroment))
	if err := d.Deploy(p, rev, file, force); err != nil {
		Critical(err.Error(), "project", p, "revision", rev)
		return err
	}

	return nil
}

func (p *Project) Test(enviroment string) {
	if p.TestCommand == "" {
		Warning("No Test command", "project", p)
		return
	}

	Info("Executing Test command", "project", p, "script", p.TestCommand)
	json, err := json.Marshal(p)
	if err != nil {
		Critical(err.Error(), "project", p)
	}

	cmd := command.NewCommand(fmt.Sprintf("%s %s %s", p.TestCommand, enviroment, json))
	if err := cmd.Run(); err != nil {
		Critical(err.Error(), "project", p, "script", p.TestCommand)
	}

	if err := cmd.Wait(); err != nil {
		Critical(err.Error(), "project", p, "script", p.TestCommand)
	}
	response := cmd.GetResponse()

	if response.Failed {
		fmt.Printf("stdout>\n%s", response.Stdout)
		fmt.Printf("stderr>\n%s", response.Stderr)
		Critical("Test command failed",
			"project", p,
			"script", p.TestCommand,
			"exitcode", response.ExitCode,
			"elapsed", response.RealTime,
		)
	} else {
		Info("Test command executed successfully",
			"project", p,
			"script", p.TestCommand,
			"exitcode", response.ExitCode,
			"elapsed", response.RealTime,
		)
	}
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
