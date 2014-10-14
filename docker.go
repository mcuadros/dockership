package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"time"

	"github.com/fsouza/go-dockerclient"
)

type Docker struct {
	client *docker.Client
}

func NewDocker(endpoint string) *Docker {
	c, _ := docker.NewClient(endpoint)

	return &Docker{client: c}
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
