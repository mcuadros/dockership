package main

import (
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/stevedomin/termtable"
)

type CmdStatus struct{ cmd }

func NewCmdStatus() (cli.Command, error) {
	return &CmdStatus{}, nil
}

func (c *CmdStatus) Run(args []string) int {
	c.buildFlags(c)
	if err := c.parse(args); err != nil {
		return 1
	}

	table := termtable.NewTable(nil, &termtable.TableOptions{Padding: 3})
	table.SetHeader([]string{"Environment", "Project", "Last Commit", "Containers", "Status"})

	for name, p := range c.config.Projects {
		if c.project != "" && c.project != name {
			continue
		}

		sl, err := p.Status()
		if len(err) != 0 {
			for _, e := range err {
				table.AddRow([]string{"-", p.String(), "-", "-", e.Error()})
			}
			continue
		}

		for _, s := range sl {
			status := "Down"
			if len(s.RunningContainers) > 0 {
				status = s.RunningContainers[0].Status
			}

			table.AddRow([]string{
				s.Environment.String(),
				p.String(),
				s.LastRevision.GetShort(),
				fmt.Sprintf("%d", len(s.Containers)),
				status,
			})
		}
	}

	fmt.Println(table.Render())
	return 0
}

func (c *CmdStatus) Synopsis() string {
	return "Prints the status from the projects."
}

func (c *CmdStatus) Help() string {
	helpText := `
Usage: dockership status [options]
  Prints the status from the projects.

Options:
  -project=""                Return the status only from this project.

`

	return strings.TrimSpace(helpText)
}
