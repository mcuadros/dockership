package main

import (
	"fmt"
	"os"

	"github.com/mcuadros/dockership/cli"
	"github.com/mcuadros/dockership/core"

	mcli "github.com/mitchellh/cli"
)

var VERSION string
var BUILD_DATE string

func main() {
	c := mcli.NewCLI("dockership", fmt.Sprintf("%s / %s", VERSION, BUILD_DATE))
	c.Args = os.Args[1:]
	c.Commands = map[string]mcli.CommandFactory{
		"status":     cli.NewCmdStatus,
		"deploy":     cli.NewCmdDeploy,
		"containers": cli.NewCmdContainers,
	}

	exitStatus, err := c.Run()
	if err != nil {
		core.Critical(err.Error())
	}

	os.Exit(exitStatus)
}
