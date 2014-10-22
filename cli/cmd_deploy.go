package main

import (
	"fmt"
	"strings"

	. "github.com/mcuadros/dockership/logger"

	"github.com/mitchellh/cli"
)

type CmdDeploy struct {
	enviroment string
	force      bool
	cmd
}

func NewCmdDeploy() (cli.Command, error) {
	return &CmdDeploy{}, nil
}

func (c *CmdDeploy) parse(args []string) error {
	c.flags.StringVar(&c.enviroment, "env", "", "")
	c.flags.BoolVar(&c.force, "force", false, "")
	err := c.cmd.parse(args)

	fmt.Println(args)

	return err
}

func (c *CmdDeploy) Run(args []string) int {
	c.buildFlags(c)
	if err := c.parse(args); err != nil {
		return 1
	}

	if p, ok := c.config.Projects[c.project]; ok {
		Info("Starting deploy", "project", p, "enviroment", c.enviroment, "force", c.force)
		err := p.Deploy(c.enviroment, c.force)
		if len(err) != 0 {
			Critical(err[0].Error(), "project", c.project)
			return 1
		}

		Info("Deploy success", "project", p, "enviroment", c.enviroment)
		return 0
	}

	Critical("Unable to find project", "project", c.project)

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
