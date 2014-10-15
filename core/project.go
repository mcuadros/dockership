package core

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/docker/docker/pkg/units"
	"github.com/docker/docker/utils"
)

type Project struct {
	GithubToken       string
	DockerEndPoint    string
	Owner             string
	Repository        string
	Branch            string `default:"master"`
	Dockerfile        string `default:"Dockerfile"`
	DockerfileContent []byte
}

func (p *Project) Deploy() error {
	c := NewGithub(p.GithubToken)
	file, commit, _ := c.GetDockerFile(p.Owner, p.Repository, p.Branch, p.Dockerfile)

	d := NewDocker(p.DockerEndPoint)
	if err := d.Deploy(p.Owner, p.Repository, commit, file); err != nil {
		return err
	}

	return nil
}

func (p *Project) Status() error {
	d := NewDocker(p.DockerEndPoint)
	l, err := d.ListContainers(p.Owner, p.Repository)
	if err != nil {
		return err
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

	return nil
}
