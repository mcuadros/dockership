package core

import (
	. "github.com/mcuadros/dockership/logger"

	"code.google.com/p/gcfg"
	"github.com/mcuadros/go-defaults"
)

type Config struct {
	Global struct {
		UseShortRevisions bool `default:"true"`
		GithubToken       string
	}

	Projects    map[string]*Project    `gcfg:"Project"`
	Enviroments map[string]*Enviroment `gcfg:"Enviroment"`
}

func (c *Config) LoadFile(filename string) error {
	err := gcfg.ReadFileInto(c, filename)
	if err != nil {
		return err
	}

	c.loadProjects()
	c.loadEnviroments()
	return nil
}

func (c *Config) loadProjects() {
	defaults.SetDefaults(c)
	for name, p := range c.Projects {
		p.Name = name
		defaults.SetDefaults(p)
		if p.GithubToken == "" {
			p.GithubToken = c.Global.GithubToken
		}

		p.UseShortRevisions = c.Global.UseShortRevisions
	}
}

func (c *Config) loadEnviroments() {
	for _, p := range c.Projects {
		p.Enviroments = make(map[string]*Enviroment, 0)
		for _, e := range p.EnviromentNames {
			p.Enviroments[e] = c.mustGetEnviroment(e)
		}
	}
}

func (c *Config) mustGetEnviroment(name string) *Enviroment {
	if e, ok := c.Enviroments[name]; ok {
		defaults.SetDefaults(e)
		e.Name = name
		return e
	}

	Critical("Undefined enviroment", "enviroment", name)
	return nil
}
