package core

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/docker/docker/pkg/units"
	"github.com/docker/docker/utils"
	"gopkg.in/inconshreveable/log15.v2"
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
	log15.Warn("Retrieving last dockerfile ...", "project", p)

	c := NewGithub(p.GithubToken)
	file, commit, err := c.GetDockerFile(p)
	if err != nil {
		log15.Error(err.Error(), "project", p)
	}

	d := NewDocker(p.DockerEndPoint)
	if err := d.Deploy(p, commit, file); err != nil {
		log15.Error(err.Error(), "project", p, "commit", commit)
		return err
	}

	return nil
}

type ProjectStatus struct {
	LastCommit        string
	RunningContainers []*Container
	Containers        []*Container
}

func (p *Project) Status() (*ProjectStatus, error) {
	s := &ProjectStatus{}

	c := NewGithub(p.GithubToken)
	if commit, err := c.GetLastCommit(p); err != nil {
		return nil, err
	} else {
		s.LastCommit = commit
	}

	d := NewDocker(p.DockerEndPoint)
	if l, err := d.ListContainers(p); err != nil {
		return nil, err
	} else {
		s.Containers = l
	}

	s.RunningContainers = make([]*Container, 0)
	for _, c := range s.Containers {
		if c.IsRunning() {
			s.RunningContainers = append(s.RunningContainers, c)
		}
	}

	return s, nil
}

func (p *Project) List() error {
	d := NewDocker(p.DockerEndPoint)
	l, err := d.ListContainers(p)
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

func (p *Project) String() string {
	return fmt.Sprintf("%s/%s@%s", p.Owner, p.Repository, p.Branch)
}
