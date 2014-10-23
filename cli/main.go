package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

	"github.com/mitchellh/cli"
)

var VERSION string
var BUILD_DATE string

const DEFAULT_CONFIG = "config.ini"

func main() {
	c := cli.NewCLI("dockership", fmt.Sprintf("%s / %s", VERSION, BUILD_DATE))
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"status":     NewCmdStatus,
		"deploy":     NewCmdDeploy,
		"containers": NewCmdContainers,
	}

	exitStatus, err := c.Run()
	if err != nil {
		Critical(err.Error())
	}

	os.Exit(exitStatus)
}

type cmd struct {
	configFile string
	project    string
	config     core.Config
	flags      *flag.FlagSet
}

func (c *cmd) loadConfig() {
	if err := c.config.LoadFile(c.configFile); err != nil {
		Critical(err.Error(), "file", c.configFile)
	}
}

func (c *cmd) buildFlags(child cli.Command) {
	c.flags = flag.NewFlagSet("set", flag.ContinueOnError)
	c.flags.StringVar(&c.configFile, "config", DEFAULT_CONFIG, "")
	c.flags.StringVar(&c.project, "project", "", "")
	c.flags.Usage = func() { child.Help() }
}

func (c *cmd) parse(args []string) error {
	if err := c.flags.Parse(args); err != nil {
		return err
	}

	c.loadConfig()

	if _, ok := c.config.Projects[c.project]; !ok {
		Critical("Unknown project", "project", c.project)
	}

	return nil
}
