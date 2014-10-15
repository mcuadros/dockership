package main

import (
	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

	"github.com/mitchellh/cli"
)

type CmdDeploy struct{}

func NewCmdDeploy() (cli.Command, error) {
	return &CmdDeploy{}, nil
}

func (c *CmdDeploy) Help() string {
	return "Help"
}

func (c *CmdDeploy) Synopsis() string {
	return "Synopsis"
}

func (c *CmdDeploy) Run(args []string) int {
	var config core.Config
	if err := config.LoadFile("config.ini"); err != nil {
		Critical(err.Error())
	}

	for _, p := range config.Project {
		err := p.Deploy()
		if err != nil {
			return 1
		}
	}

	return 0
}
