package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/docker/docker/pkg/units"
	"github.com/docker/docker/utils"
	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("app", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"foo": NewCmdStatus,
	}

	exitStatus, err := c.Run()
	if err != nil {
		Error(err.Error())
	}

	os.Exit(exitStatus)
}

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
	var config Config
	config.LoadFile("config.ini")

	for _, p := range config.Project {
		p.Status()
	}

	return 0
}

type Project struct {
	GithubToken       string
	DockerEndPoint    string
	Owner             string
	Repository        string
	Branch            string `default:"master"`
	Dockerfile        string `default:"Dockerfile"`
	DockerfileContent []byte
}

func (p *Project) Deploy() {
	c := NewGithub(p.GithubToken)
	file, commit, _ := c.GetDockerFile(p.Owner, p.Repository, p.Branch, p.Dockerfile)

	d := NewDocker(p.DockerEndPoint)
	if err := d.Deploy(p.Owner, p.Repository, commit, file); err != nil {
		Critical(err.Error())
	}
}

func (p *Project) Status() {
	d := NewDocker(p.DockerEndPoint)
	l, err := d.ListContainers(p.Owner, p.Repository)
	if err != nil {
		Critical(err.Error())
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	fmt.Fprint(w, "REPOSITORY\tCOMMIT\tCONTAINER ID\tCOMMAND\tCREATED\tSTATUS\tPORTS\n")

	for _, c := range l {
		owner, repository, commit := c.Image.GetInfo()
		fmt.Fprintf(w, "%s/%s\t%s\t%s\t%s\t%s ago\t%s\t%s\t\n",
			owner,
			repository,
			utils.TruncateID(commit),
			utils.TruncateID(c.ID),
			c.Command,
			units.HumanDuration(time.Now().UTC().Sub(time.Unix(c.Created, 0))),
			c.Status,
			c.GetPorts(),
		)

	}

	w.Flush()
}
