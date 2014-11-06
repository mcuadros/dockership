package cli

import (
	"bytes"
	"strings"

	"github.com/mcuadros/dockership/core"

	"github.com/mitchellh/cli"
)

type CmdDeploy struct {
	environment string
	force       bool
	cmd
}

func NewCmdDeploy() (cli.Command, error) {
	return &CmdDeploy{}, nil
}

func (c *CmdDeploy) parse(args []string) error {
	c.flags.StringVar(&c.environment, "env", "", "")
	c.flags.BoolVar(&c.force, "force", false, "")
	err := c.cmd.parse(args)

	return err
}

func (c *CmdDeploy) Run(args []string) int {
	c.buildFlags(c)
	if err := c.parse(args); err != nil {
		return 1
	}

	if p, ok := c.config.Projects[c.project]; ok {
		core.Info("Starting deploy", "project", p, "environment", c.environment, "force", c.force)
		input := bytes.NewBuffer(nil)
		err := p.Deploy(c.environment, input, c.force)
		if len(err) != 0 {
			core.Critical(err[0].Error(), "project", c.project)
			return 1
		}

		core.Info("Deploy success", "project", p, "environment", c.environment)
		return 0
	}

	core.Critical("Unable to find project", "project", c.project)

	return 1
}

func (c *CmdDeploy) Synopsis() string {
	return "Deploy a project in the target environment."
}

func (c *CmdDeploy) Help() string {
	helpText := `
Usage: dockership deploy [options]
  Deploy a project in the target environment. The dockership will search for the
  last commit at the gived repository, retrieving the Dockerfile. This Dockerfile
  will be used for create a new image and launching a new container. Cleaning
  the old images in the process.


Options:
  -project=""                Just deploy the given project.
  -env=""                    Target environment.
  -force                     Deploy even a container is allready running.

`
	return strings.TrimSpace(helpText)
}
