package config

import (
	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

	"code.google.com/p/gcfg"
	"github.com/mcuadros/go-defaults"
)

const DEFAULT_CONFIG = "config.ini"

type Config struct {
	Global struct {
		UseShortRevisions bool `default:"true"`
		GithubToken       string
	}
	HTTP struct {
		Listen             string `default:":8080"`
		GithubID           string
		GithubSecret       string
		GithubOrganization string
		GithubRedirectURL  string
	}
	Projects    map[string]*core.Project    `gcfg:"Project"`
	Enviroments map[string]*core.Enviroment `gcfg:"Enviroment"`
}

func (c *Config) LoadFile(filename string) error {
	err := gcfg.ReadFileInto(c, filename)
	if err != nil {
		return err
	}

	defaults.SetDefaults(c)
	c.LoadProjects()
	c.LoadEnviroments()
	return nil
}

func (c *Config) LoadProjects() {
	defaults.SetDefaults(c)
	for name, p := range c.Projects {
		p.Name = name
		defaults.SetDefaults(p)
		if p.GithubToken == "" {
			p.GithubToken = c.Global.GithubToken
		}

		p.UseShortRevisions = c.Global.UseShortRevisions
		p.LinkedBy = make([]*core.Project, 0)
	}
}

func (c *Config) LoadEnviroments() {
	for _, p := range c.Projects {
		p.Enviroments = make(map[string]*core.Enviroment, 0)
		for _, e := range p.EnviromentNames {
			p.Enviroments[e] = c.mustGetEnviroment(p, e)
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

func (c *Config) mustGetEnviroment(p *core.Project, name string) *core.Enviroment {
	if e, ok := c.Enviroments[name]; ok {
		defaults.SetDefaults(e)
		e.Name = name
		return e
	}

	Critical("Undefined enviroment", "enviroment", name, "project", p)
	return nil
}

func (c *Config) mustGetProject(p *core.Project, name string) *core.Project {
	if e, ok := c.Projects[name]; ok {
		defaults.SetDefaults(e)
		return e
	}

	Critical("Undefined project", "project", name, "project", p)
	return nil
}
