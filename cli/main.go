package main

import (
	"os"

	. "github.com/mcuadros/dockership/logger"

	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("app", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"status": NewCmdStatus,
		"deploy": NewCmdDeploy,
	}

	exitStatus, err := c.Run()
	if err != nil {
		Critical(err.Error())
	}

	os.Exit(exitStatus)
}
