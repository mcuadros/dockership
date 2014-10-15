package main

import (
	"flag"
	"strings"

	"github.com/mcuadros/dockership/core"

	"github.com/mitchellh/cli"
)

type CmdDeploy struct {
	config *core.Config
}

func NewCmdDeploy() (cli.Command, error) {
	var config core.Config
	config.LoadFile("config.ini")

	return &CmdDeploy{config: &config}, nil
}

func (c *CmdDeploy) Run(args []string) int {
	var project string
	cmdFlags := flag.NewFlagSet("deploy", flag.ContinueOnError)
	cmdFlags.StringVar(&project, "project", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	for name, p := range c.config.Project {
		if project != "" && project != name {
			continue
		}

		err := p.Deploy()
		if err != nil {
			return 1
		}
	}

	return 0
}

func (c *CmdDeploy) Synopsis() string {
	return "Deploy a project in the target enviroment."
}

func (c *CmdDeploy) Help() string {
	helpText := `
Usage: dockership deploy [options]
  Deploy a project in the target enviroment. The dockership will search for the
  last commit at the gived repository, retrieving the Dockerfile. This Dockerfile
  will be used for create a new image and launching a new container. Cleaning
  the old images in the process.


Options:
  -project=""                Just deploy the given project.

`

	return strings.TrimSpace(helpText)
}
