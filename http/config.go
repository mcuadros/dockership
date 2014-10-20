package main

import (
	"github.com/mcuadros/dockership/core"

	"code.google.com/p/gcfg"
	"github.com/mcuadros/go-defaults"
)

type config struct {
	HTTP struct {
		Listen string `default:":8080"`
	}
	core.Config
}

func (c *config) LoadFile(filename string) error {
	err := gcfg.ReadFileInto(c, filename)
	if err != nil {
		return err
	}

	defaults.SetDefaults(c)
	c.LoadProjects()
	c.LoadEnviroments()
	return nil
}
