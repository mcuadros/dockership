package main

import (
	"fmt"

	"github.com/mcuadros/dockership/core"

	"github.com/mitchellh/cli"
	"github.com/stevedomin/termtable"
)

type CmdStatus struct{}

func NewCmdStatus() (cli.Command, error) {
	return &CmdStatus{}, nil
}

func (c *CmdStatus) Help() string {
	return "Help"
}

func (c *CmdStatus) Synopsis() string {
	return "Synopsis"
}

func (c *CmdStatus) Run(args []string) int {
	var config core.Config
	config.LoadFile("config.ini")

	table := termtable.NewTable(nil, &termtable.TableOptions{Padding: 3})
	table.SetHeader([]string{"Project", "Last Commit", "Containers", "Status"})

	for _, p := range config.Project {
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
