package core

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"gopkg.in/inconshreveable/log15.v2"
)

var statusUp = regexp.MustCompile("^Up (.*)")
var imageIdRe = regexp.MustCompile("^(.*)/(.*):(.*)")

type Docker struct {
	client *docker.Client
}

func NewDocker(endpoint string) *Docker {
	log15.Debug("Connected to docker endpoint", "endpoint", endpoint)

	c, _ := docker.NewClient(endpoint)

	return &Docker{client: c}
}

func (d *Docker) Deploy(p *Project, commit string, dockerfile []byte) error {
	log15.Info("Deploying dockerfile", "project", p, "commit", commit)
	if err := d.Clean(p, commit); err != nil {
		return err
	}

	if err := d.BuildImage(p, commit, dockerfile); err != nil {
		return err
	}

	return d.Run(p, commit)
}

func (d *Docker) Clean(p *Project, commit string) error {
	l, err := d.ListContainers(p)
	if err != nil {
		return err
	}

	for _, c := range l {
		if c.IsRunning() && c.Image.IsCommit(commit) {
			return errors.New("Current commit is already running")
		}
	}

	log15.Info("Cleaning all containers", "project", p)
	for _, c := range l {
		log15.Info("Killing and removing image", "project", p, "image", c.GetShortId())
		err := d.killAndRemove(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Docker) ListContainers(p *Project) ([]*Container, error) {
	log15.Debug("Retrieving current containers", "project", p)

	l, err := d.client.ListContainers(docker.ListContainersOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	r := make([]*Container, 0)
	for _, c := range l {
		i := ImageId(c.Image)
		if i.BelongsTo(p) {
			r = append(r, &Container{Image: i, APIContainers: c})
		}
	}

	return r, nil
}

func (d *Docker) killAndRemove(c *Container) error {
	kopts := docker.KillContainerOptions{ID: c.ID}
	if err := d.client.KillContainer(kopts); err != nil {
		return err
	}

	ropts := docker.RemoveContainerOptions{ID: c.ID}
	if err := d.client.RemoveContainer(ropts); err != nil {
		return err
	}

	return nil
}

func (d *Docker) BuildImage(p *Project, commit string, dockerfile []byte) error {
	log15.Info("Building image", "project", p, "commit", commit)

	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	d.buildTar(dockerfile, inputbuf)

	image := d.getImageName(p, commit)
	opts := docker.BuildImageOptions{
		Name:         string(image),
		InputStream:  inputbuf,
		OutputStream: outputbuf,
	}

	return d.client.BuildImage(opts)
}

func (d *Docker) Run(p *Project, commit string) error {
	log15.Debug("Creating container from image", "project", p, "commit", commit)
	c, err := d.createContainer(d.getImageName(p, commit))
	if err != nil {
		return err
	}

	log15.Info("Running new container",
		"project", p,
		"commit", commit,
		"container", c.GetShortId(),
	)

	return d.startContainer(c)
}

func (d *Docker) getImageName(p *Project, commit string) ImageId {
	return ImageId(fmt.Sprintf("%s/%s:%s", p.Owner, p.Repository, commit))
}

func (d *Docker) createContainer(image ImageId) (*Container, error) {
	c, err := d.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: string(image),
		},
	})

	if err != nil {
		return nil, err
	}

	return &Container{Image: image, APIContainers: docker.APIContainers{ID: c.ID}}, nil
}

func (d *Docker) startContainer(c *Container) error {
	return d.client.StartContainer(c.ID, &docker.HostConfig{
		PortBindings: map[docker.Port][]docker.PortBinding{
			"80/tcp": []docker.PortBinding{docker.PortBinding{
				HostIp:   "0.0.0.0",
				HostPort: "212",
			}},
		},
	})
}

func (d *Docker) buildTar(dockerfile []byte, buf *bytes.Buffer) *tar.Writer {
	t := time.Now()

	tr := tar.NewWriter(buf)
	tr.WriteHeader(&tar.Header{
		Name:       "Dockerfile",
		Size:       int64(len(dockerfile)),
		ModTime:    t,
		AccessTime: t,
		ChangeTime: t,
	})

	tr.Write(dockerfile)
	tr.Close()

	return tr
}

type ImageId string

func (i ImageId) BelongsTo(p *Project) bool {
	return strings.HasPrefix(string(i), fmt.Sprintf("%s/%s", p.Owner, p.Repository))
}

func (i ImageId) IsCommit(commit string) bool {
	return strings.HasSuffix(string(i), commit)
}

func (i ImageId) GetInfo() (owner, repository, commit string) {
	m := imageIdRe.FindStringSubmatch(string(i))
	owner, repository, commit = m[1], m[2], m[3]

	return
}

type Container struct {
	Image ImageId
	docker.APIContainers
}

func (c *Container) IsRunning() bool {
	return statusUp.MatchString(c.Status)
}

func (c *Container) GetShortId() string {
	shortLen := 12
	if len(c.ID) < shortLen {
		shortLen = len(c.ID)
	}

	return c.ID[:shortLen]
}

func (c *Container) GetPorts() string {
	result := []string{}
	for _, port := range c.Ports {
		if port.IP == "" {
			result = append(result, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		} else {
			result = append(result, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
		}
	}
	return strings.Join(result, ", ")
}
