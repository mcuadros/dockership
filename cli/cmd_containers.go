package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/stevedomin/termtable"
)

type CmdContainers struct{ cmd }

func NewCmdContainers() (cli.Command, error) {
	return &CmdContainers{}, nil
}

func (c *CmdContainers) Run(args []string) int {
	c.buildFlags(c)
	if err := c.parse(args); err != nil {
		return 1
	}

	table := termtable.NewTable(nil, &termtable.TableOptions{Padding: 3})
	table.SetHeader([]string{"Environment", "Repository", "Commit", "Container ID", "Created", "Command", "Status", "Ports"})

	for name, p := range c.config.Projects {
		if c.project != "" && c.project != name {
			continue
		}

		l, err := p.ListContainers()
		if len(err) != 0 {
			continue
		}

		for _, c := range l {
			table.AddRow([]string{
				c.DockerEndPoint,
				p.String(),
				c.Image.GetRevisionString(),
				c.GetShortId(),
				c.Command,
				HumanDuration(time.Now().UTC().Sub(time.Unix(c.Created, 0))),
				c.Status,
				c.GetPortsString(),
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
Usage: dockership containers [options]
  List the containers deployed by this tool.


Options:
  -project=""                Just show the containers for this project.

`

	return strings.TrimSpace(helpText)
}
