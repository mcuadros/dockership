package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mcuadros/dockership/core"

	"github.com/mitchellh/cli"
	"github.com/stevedomin/termtable"
)

type CmdStatus struct {
	config *core.Config
}

func NewCmdStatus() (cli.Command, error) {
	var config core.Config
	config.LoadFile("config.ini")

	return &CmdStatus{config: &config}, nil
}

func (c *CmdStatus) Run(args []string) int {
	var project string
	cmdFlags := flag.NewFlagSet("status", flag.ContinueOnError)
	cmdFlags.StringVar(&project, "project", "", "")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	table := termtable.NewTable(nil, &termtable.TableOptions{Padding: 3})
	table.SetHeader([]string{"Project", "Last Commit", "Containers", "Status"})

	for name, p := range c.config.Project {
		if project != "" && project != name {
			continue
		}

		s, err := p.Status()
		if err != nil {
			table.AddRow([]string{p.String(), "-", "-", err.Error()})
			continue

		}

		status := "Down"
		if len(s.RunningContainers) > 0 {
			status = s.RunningContainers[0].Status
		}

		table.AddRow([]string{
			p.String(),
			s.LastCommit,
			fmt.Sprintf("%d", len(s.Containers)),
			status,
		})

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
