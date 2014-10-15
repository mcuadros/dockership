package main

import (
	"github.com/mcuadros/dockership/core"

	"github.com/mitchellh/cli"
)

type CmdStatus struct{}

func NewCmdStatus() (cli.Command, error) {
	return &CmdStatus{}, nil
}

func (c *CmdStatus) Help() string {
	return "Help"
}

func (c *CmdStatus) Synopsis() string {
	return "Synopsis"
}

func (c *CmdStatus) Run(args []string) int {
	var config core.Config
	config.LoadFile("config.ini")

	for _, p := range config.Project {
		p.Status()
	}

	return 0
}
