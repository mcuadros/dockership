package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mcuadros/dockership/core"

	"github.com/docker/docker/pkg/units"
	"github.com/mitchellh/cli"
	"github.com/stevedomin/termtable"
)

type CmdContainers struct{}

func NewCmdContainers() (cli.Command, error) {
	return &CmdContainers{}, nil
}

func (c *CmdContainers) Run(args []string) int {
	var config core.Config
	config.LoadFile("config.ini")

	table := termtable.NewTable(nil, &termtable.TableOptions{Padding: 3})
	table.SetHeader([]string{"Repository", "Commit", "Container ID", "Created", "Command", "Status", "Ports"})

	for _, p := range config.Project {
		l, err := p.List()
		if err != nil {
			continue
		}

		for _, c := range l {
			_, _, commit := c.Image.GetInfo()
			table.AddRow([]string{
				p.String(),
				commit.GetShort(),
				c.GetShortId(),
				c.Command,
				units.HumanDuration(time.Now().UTC().Sub(time.Unix(c.Created, 0))),
				c.Status,
				c.GetPorts(),
			})
		}
	}

	fmt.Println(table.Render())
	return 0
}

func (c *CmdContainers) Synopsis() string {
	return "List the containers deployed by this tool."
}

func (c *CmdContainers) Help() string {
	helpText := `
Usage: dockership deploy [options]
  List the containers deployed by this tool.


Options:
  -project=""                Just show the containers for this project.

`

	return strings.TrimSpace(helpText)
}
