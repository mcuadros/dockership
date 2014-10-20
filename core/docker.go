package core

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	. "github.com/mcuadros/dockership/logger"
)

type Docker struct {
	enviroment *Enviroment
	client     *docker.Client
}

func NewDocker(enviroment *Enviroment) *Docker {
	Debug("Connected to docker", "enviroment", enviroment)
	c, _ := docker.NewClient(enviroment.DockerEndPoint)

	return &Docker{client: c, enviroment: enviroment}
}

func (d *Docker) Deploy(p *Project, rev Revision, dockerfile []byte, force bool) error {
	Info("Deploying dockerfile", "project", p, "revision", rev)
	if err := d.Clean(p); err != nil {
		return err
	}

	if err := d.BuildImage(p, rev, dockerfile); err != nil {
		return err
	}

	return d.Run(p, rev)
}

func (d *Docker) Clean(p *Project) error {
	l, err := d.ListContainers(p)
	if err != nil {
		return err
	}

	count := len(l)
	if count < 1 {
		return nil
	}

	keep := d.enviroment.History
	if keep < 1 {
		keep = 1
	}

	Info("Removing old containers", "project", p, "count", count)
	for _, c := range l[:count-keep] {
		Info("Killing and removing image", "project", p, "container", c.GetShortId())
		err := d.killAndRemove(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Docker) ListContainers(p *Project) ([]*Container, error) {
	Debug("Retrieving current containers", "project", p)

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
			r = append(r, &Container{
				Image:         i,
				APIContainers: c,
				Enviroment:    d.enviroment,
			})
		}
	}

	sort.Sort(SortByCreated(r))

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

	if err := d.client.RemoveImage(string(c.Image)); err != nil {
		return err
	}

	return nil
}

func (d *Docker) BuildImage(p *Project, rev Revision, dockerfile []byte) error {
	Info("Building image", "project", p, "revision", rev)

	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	outputbuf.WriteTo(os.Stdout)

	if err := d.buildTar(p, dockerfile, inputbuf); err != nil {
		return err
	}

	image := d.getImageName(p, rev)
	opts := docker.BuildImageOptions{
		Name:           string(image),
		NoCache:        p.NoCache,
		RmTmpContainer: p.NoCache,
		InputStream:    inputbuf,
		OutputStream:   outputbuf,
	}

	return d.client.BuildImage(opts)
}

func (d *Docker) Run(p *Project, rev Revision) error {
	Debug("Creating container from image", "project", p, "revision", rev)
	c, err := d.createContainer(d.getImageName(p, rev))
	if err != nil {
		return err
	}

	Info("Running new container",
		"project", p,
		"revision", rev,
		"image", c.Image,
		"container", c.GetShortId(),
	)

	return d.startContainer(p, c)
}

func (d *Docker) getImageName(p *Project, rev Revision) ImageId {
	c := rev.String()
	if p.UseShortRevisions {
		c = rev.GetShort()
	}

	info := p.Repository.Info()
	return ImageId(fmt.Sprintf("%s/%s:%s", info.Username, info.Name, c))
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

func (d *Docker) startContainer(p *Project, c *Container) error {
	ports, err := d.formatPorts(p.Ports)
	if err != nil {
		return err
	}

	return d.client.StartContainer(c.ID, &docker.HostConfig{
		PortBindings: ports,
	})
}

func (d *Docker) formatPorts(ports []string) (map[docker.Port][]docker.PortBinding, error) {
	r := make(map[docker.Port][]docker.PortBinding, 0)
	for _, p := range ports {
		guest, host, err := d.formatPort(p)
		if err != nil {
			return nil, err
		}

		if _, ok := r[guest]; !ok {
			r[guest] = make([]docker.PortBinding, 0)
		}

		r[guest] = append(r[guest], host)
	}

	return r, nil
}

// <host_interface>:<host_port>:<container_port>/<proto>
func (d *Docker) formatPort(port string) (guest docker.Port, host docker.PortBinding, err error) {
	p1 := strings.SplitN(port, "/", 2)
	p2 := strings.SplitN(p1[0], ":", 3)

	if len(p1) != 2 || len(p2) != 3 {
		err = errors.New(fmt.Sprintf("Malformed port %q", port))
		return
	}

	guest = docker.Port(fmt.Sprintf("%s/%s", p2[2], p1[1]))
	host = docker.PortBinding{
		HostIp:   p2[0],
		HostPort: p2[1],
	}

	return
}

func (d *Docker) buildTar(p *Project, dockerfile []byte, buf *bytes.Buffer) error {
	t := time.Now()

	tr := tar.NewWriter(buf)
	tr.WriteHeader(&tar.Header{
		Name:       "Dockerfile",
		Size:       int64(len(dockerfile)),
		ModTime:    t,
		AccessTime: t,
		ChangeTime: t,
	})

	if _, err := tr.Write(dockerfile); err != nil {
		return err
	}

	for _, file := range p.Files {
		if err := d.addFileToTar(file, tr); err != nil {
			return err
		}
	}

	tr.Close()
	return nil
}

func (d *Docker) addFileToTar(file string, tr *tar.Writer) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	fInfo, err := os.Lstat(file)
	if err != nil {
		return err
	}

	h, err := tar.FileInfoHeader(fInfo, "")
	h.Name = path.Base(file)
	if err != nil {
		return err
	}

	if err := tr.WriteHeader(h); err != nil {
		return err
	}

	if _, err := tr.Write(content); err != nil {
		return err
	}

	return nil
}
