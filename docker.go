package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var statusUp = regexp.MustCompile("^Up (.*)")
var imageIdRe = regexp.MustCompile("^(.*)/(.*):(.*)")

type ImageId string

func (i ImageId) BelongsTo(owner, repository string) bool {
	return strings.HasPrefix(string(i), fmt.Sprintf("%s/%s", owner, repository))
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

type Docker struct {
	client *docker.Client
}

func NewDocker(endpoint string) *Docker {
	c, _ := docker.NewClient(endpoint)

	return &Docker{client: c}
}

func (d *Docker) Deploy(owner, repository, commit string, dockerfile []byte) error {
	if err := d.Clean(owner, repository, commit); err != nil {
		return err
	}

	if err := d.BuildImage(owner, repository, commit, dockerfile); err != nil {
		return err
	}

	return d.Run(owner, repository, commit)
}

func (d *Docker) Clean(owner, repository, commit string) error {
	l, err := d.ListContainers(owner, repository)
	if err != nil {
		return err
	}

	for _, c := range l {
		if c.IsRunning() && c.Image.IsCommit(commit) {
			return errors.New("Current commits is already running")
		}
	}

	for _, c := range l {
		err := d.KillAndRemove(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Docker) ListContainers(owner, repository string) ([]*Container, error) {
	l, err := d.client.ListContainers(docker.ListContainersOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	r := make([]*Container, 0)
	for _, c := range l {
		i := ImageId(c.Image)
		if i.BelongsTo(owner, repository) {
			r = append(r, &Container{Image: i, APIContainers: c})
		}
	}

	return r, nil
}

func (d *Docker) KillAndRemove(c *Container) error {
	fmt.Println(c.Image)

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

func (d *Docker) BuildImage(owner, repository, commit string, dockerfile []byte) error {
	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	d.buildTar(dockerfile, inputbuf)

	image := d.getImageName(owner, repository, commit)
	opts := docker.BuildImageOptions{
		Name:         image,
		InputStream:  inputbuf,
		OutputStream: outputbuf,
	}

	return d.client.BuildImage(opts)
}

func (d *Docker) Run(owner, repository, commit string) error {
	c, err := d.createContainer(d.getImageName(owner, repository, commit))
	if err != nil {
		return err
	}

	return d.startContainer(c)
}

func (d *Docker) getImageName(owner, repository, commit string) string {
	return fmt.Sprintf("%s/%s:%s", owner, repository, commit)
}

func (d *Docker) createContainer(image string) (*docker.Container, error) {
	return d.client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image,
		},
	})
}

func (d *Docker) startContainer(c *docker.Container) error {
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
