package main

import (
	"os"

	. "github.com/mcuadros/dockership/logger"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("dockership", "0.0.1")
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
