package core

import (
	"code.google.com/p/gcfg"
	"github.com/mcuadros/go-defaults"
)

type Config struct {
	Main struct {
		UseShortCommits bool `default:"true"`
		GithubToken     string
		DockerEndPoint  string
	}
	Project map[string]*Project
}

func (c *Config) LoadFile(filename string) error {
	err := gcfg.ReadFileInto(c, filename)
	if err != nil {
		return err
	}

	c.loadDefaults()
	return nil
}

func (c *Config) loadDefaults() {
	defaults.SetDefaults(c)
	for _, p := range c.Project {
		defaults.SetDefaults(p)
		if p.GithubToken == "" {
			p.GithubToken = c.Main.GithubToken
		}

		if p.DockerEndPoint == "" {
			p.DockerEndPoint = c.Main.DockerEndPoint
		}
	}
}
