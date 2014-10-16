package main

import (
	"flag"
	"strings"

	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

	"github.com/mitchellh/cli"
)

type CmdDeploy struct {
	config *core.Config
}

func NewCmdDeploy() (cli.Command, error) {
	var config core.Config
	if err := config.LoadFile("config.ini"); err != nil {
		Critical(err.Error(), "file", "config.ini")
	}

	return &CmdDeploy{config: &config}, nil
}

func (c *CmdDeploy) Run(args []string) int {
	var project, enviroment string
	var force bool
	cmdFlags := flag.NewFlagSet("deploy", flag.ContinueOnError)
	cmdFlags.StringVar(&project, "project", "", "")
	cmdFlags.StringVar(&enviroment, "env", "", "")
	cmdFlags.BoolVar(&force, "force", false, "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if p, ok := c.config.Projects[project]; ok {
		Info("Starting deploy", "project", p, "enviroment", enviroment, "force", force)
		err := p.Deploy(force, enviroment)
		if err != nil {
			Critical(err.Error(), "project", project)
			return 1
		}

		p.Test(enviroment)
		Info("Deploy success", "project", p, "enviroment", enviroment)
		return 0
	}

	Critical("Unable to find project", "project", project)

	return 1
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
  -env=""                    Target enviroment.
  -force                     Deploy even a container is allready running.

`
	return strings.TrimSpace(helpText)
}
