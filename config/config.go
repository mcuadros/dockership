package config

import (
	"github.com/mcuadros/dockership/core"

	"code.google.com/p/gcfg"
	"github.com/mcuadros/go-defaults"
)

const DEFAULT_CONFIG = "/etc/dockership/dockership.conf"

type Config struct {
	Global struct {
		UseShortRevisions bool `default:"true"`
		GithubToken       string
		EtcdServers       []string `gcfg:"EtcdServer"`
	}
	HTTP struct {
		Listen             string `default:":8080"`
		GithubID           string
		GithubSecret       string
		GithubOrganization string
		GithubUsers        []string `gcfg:"GithubUser"`
		GithubRedirectURL  string
	}
	Projects     map[string]*core.Project     `gcfg:"Project"`
	Environments map[string]*core.Environment `gcfg:"Environment"`
}

func (c *Config) LoadFile(filename string) error {
	err := gcfg.ReadFileInto(c, filename)
	if err != nil {
		return err
	}

	defaults.SetDefaults(c)
	c.LoadProjects()
	c.LoadEnvironments()
	c.LinkProjectsAndEnviroments()
	return nil
}

func (c *Config) LoadProjects() {
	for name, p := range c.Projects {
		p.Name = name
		defaults.SetDefaults(p)
		if p.GithubToken == "" {
			p.GithubToken = c.Global.GithubToken
		}

		p.UseShortRevisions = c.Global.UseShortRevisions
		p.LinkedBy = make([]*core.Project, 0)
		p.TaskStatus = core.TaskStatus{}
	}
}

func (c *Config) LoadEnvironments() {
	for name, e := range c.Environments {
		e.Name = name
		defaults.SetDefaults(e)
		if e.EtcdServers == nil || len(e.EtcdServers) == 0 {
			e.EtcdServers = c.Global.EtcdServers
		}
	}
}

func (c *Config) LinkProjectsAndEnviroments() {
	for _, p := range c.Projects {
		p.Environments = make(map[string]*core.Environment, 0)
		for _, e := range p.EnvironmentNames {
			p.Environments[e] = c.mustGetEnvironment(p, e)
		}

		p.Links = make(map[string]*core.Link, 0)
		for _, l := range p.LinkNames {
			linked := c.mustGetProject(p, l.GetProjectName())
			p.Links[l.GetProjectName()] = &core.Link{
				Alias:   l.GetAlias(),
				Project: linked,
			}

			linked.LinkedBy = append(linked.LinkedBy, p)
		}
	}
}

func (c *Config) mustGetEnvironment(p *core.Project, name string) *core.Environment {
	if e, ok := c.Environments[name]; ok {
		defaults.SetDefaults(e)
		e.Name = name
		return e
	}

	core.Critical("Undefined environment", "environment", name, "project", p)
	return nil
}

func (c *Config) mustGetProject(p *core.Project, name string) *core.Project {
	if e, ok := c.Projects[name]; ok {
		defaults.SetDefaults(e)
		return e
	}

	core.Critical("Undefined project", "project", name, "project", p)
	return nil
}
