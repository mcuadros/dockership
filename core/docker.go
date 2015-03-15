package core

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

type Docker struct {
	endPoint string
	env      *Environment
	client   *docker.Client
}

func NewDocker(endPoint string, env *Environment) (*Docker, error) {
	Debug("Connected to docker", "end-point", endPoint)

	var c *docker.Client
	var err error
	if env != nil && env.CertPath != "" {
		c, err = docker.NewTLSClient(
			endPoint,
			env.CertPath+"/cert.pem",
			env.CertPath+"/key.pem",
			env.CertPath+"/ca.pem",
		)
	} else {
		c, err = docker.NewClient(endPoint)
	}

	if err != nil {
		return nil, err
	}

	return &Docker{client: c, endPoint: endPoint, env: env}, nil
}

func (d *Docker) Deploy(p *Project, rev Revision, dockerfile *Dockerfile, output io.Writer, force bool) error {
	Debug("Deploying dockerfile", "project", p, "revision", rev, "end-point", d.endPoint)

	if err := d.cleanImages(p); err != nil {
		return err
	}

	if err := d.BuildImage(p, rev, dockerfile, output); err != nil {
		return err
	}

	if err := d.cleanContainers(p); err != nil {
		return err
	}

	return d.Run(p, rev)
}

func (d *Docker) Clean(p *Project) error {
	if err := d.cleanContainers(p); err != nil {
		return err
	}

	if err := d.cleanImages(p); err != nil {
		return err
	}

	return nil
}

func (d *Docker) cleanContainers(p *Project) error {
	l, err := d.ListContainers(p)
	if err != nil {
		return err
	}

	Debug("Cleaning containers", "project", p, "count", len(l), "end-point", d.endPoint)
	for _, c := range l {
		if c.IsRunning() {
			Debug("Stoping container and image", "project", p, "container", c.GetShortId(), "end-point", d.endPoint)
			if err := d.killContainer(c); err != nil {
				return err
			}
		}

		Debug("Removing container", "project", p, "container", c.GetShortId(), "end-point", d.endPoint)
		if err := d.removeContainer(c); err != nil {
			return err
		}
	}

	return nil
}

func (d *Docker) killContainer(c *Container) error {
	kopts := docker.KillContainerOptions{ID: c.ID}
	if err := d.client.KillContainer(kopts); err != nil {
		return err
	}

	return nil
}

func (d *Docker) removeContainer(c *Container) error {
	ropts := docker.RemoveContainerOptions{ID: c.ID}
	if err := d.client.RemoveContainer(ropts); err != nil {
		return err
	}

	return nil
}

func (d *Docker) cleanImages(p *Project) error {
	l, err := d.ListImages(p)
	if err != nil {
		return err
	}

	keep := p.History
	if keep < 0 {
		keep = 0
	}

	count := len(l)
	if count < keep {
		return nil
	}

	Debug("Removing old images", "project", p, "count", count-keep, "end-point", d.endPoint)
	for _, i := range l[:count-keep] {
		Debug("Removing image", "project", p, "image", i.ID, "end-point", d.endPoint)
		if err := d.removeImage(i); err != nil {
			return err
		}
	}

	return nil
}

func (d *Docker) removeImage(i *Image) error {
	if err := d.client.RemoveImage(i.ID); err != nil {
		return err
	}

	return nil
}

func (d *Docker) ListContainers(p *Project) ([]*Container, error) {
	Debug("Retrieving current containers", "project", p, "end-point", d.endPoint)

	l, err := d.client.ListContainers(docker.ListContainersOptions{
		All: true,
	})

	if err != nil {
		return nil, err
	}

	r := make([]*Container, 0)
	for _, c := range l {
		container := &Container{
			Image:          ImageId(c.Image),
			APIContainers:  c,
			DockerEndPoint: d.endPoint,
		}

		if container.BelongsTo(p) {
			r = append(r, container)
		}
	}

	sort.Sort(ContainersByCreated(r))

	return r, nil
}

func (d *Docker) ListImages(p *Project) ([]*Image, error) {
	Debug("Retrieving current containers", "project", p, "end-point", d.endPoint)

	l, err := d.client.ListImages(docker.ListImagesOptions{All: true})

	if err != nil {
		return nil, err
	}

	r := make([]*Image, 0)
	for _, i := range l {
		image := &Image{
			APIImages:      i,
			DockerEndPoint: d.endPoint,
		}

		if image.BelongsTo(p) {
			r = append(r, image)
		}
	}

	sort.Sort(ImagesByCreated(r))

	return r, nil
}

func (d *Docker) BuildImage(
	p *Project, rev Revision, dockerfile *Dockerfile, output io.Writer,
) error {
	Debug("Building image", "project", p, "revision", rev, "end-point", d.endPoint)

	input := bytes.NewBuffer(nil)
	if err := d.buildTar(p, dockerfile.Get(), input); err != nil {
		return err
	}

	image := d.getImageName(p, rev)
	opts := docker.BuildImageOptions{
		Name:           string(image),
		NoCache:        p.NoCache,
		RmTmpContainer: p.NoCache,
		InputStream:    input,
		OutputStream:   output,
	}

	return d.client.BuildImage(opts)
}

func (d *Docker) Run(p *Project, rev Revision) error {
	Debug("Creating container from image", "project", p, "revision", rev, "end-point", d.endPoint)
	c, err := d.createContainer(p, d.getImageName(p, rev))
	if err != nil {
		return err
	}

	Info("Running new container",
		"project", p,
		"revision", rev.GetShort(),
		"container", c.GetShortId(),
		"end-point", d.endPoint,
	)

	if err := d.startContainer(p, c); err != nil {
		return err
	}

	return d.restartLinkedContainers(p)
}

func (d *Docker) getImageName(p *Project, rev Revision) ImageId {
	c := rev.String()
	if p.UseShortRevisions {
		c = rev.GetShort()
	}

	return ImageId(fmt.Sprintf("%s:%s", p.Name, c))
}

func (d *Docker) createContainer(p *Project, image ImageId) (*Container, error) {
	c, err := d.client.CreateContainer(docker.CreateContainerOptions{
		Name: p.Name,
		Config: &docker.Config{
			Image:   string(image),
			Volumes: p.Volumes,
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

	restartPolicy, err := d.formatRestartPolicy(p.Restart)
	if err != nil {
		return err
	}

	return d.client.StartContainer(c.ID, &docker.HostConfig{
		PortBindings:  ports,
		RestartPolicy: restartPolicy,
		Links:         d.formatLinks(p.Links),
		VolumesFrom:   p.VolumesFrom,
	})
}

func (d *Docker) formatLinks(links map[string]*Link) []string {
	r := make([]string, 0)
	for _, link := range links {
		r = append(r, link.String())
	}

	return r
}

func (d *Docker) formatRestartPolicy(restart string) (policy docker.RestartPolicy, err error) {
	values := strings.SplitN(restart, ":", 2)
	if values[0] == "no" || restart == "" {
		policy = docker.NeverRestart()
		return
	}

	if values[0] == "always" {
		policy = docker.AlwaysRestart()
		return
	}

	if values[0] == "on-failure" {
		var maxRetry int
		maxRetry, err = strconv.Atoi(values[1])
		if err != nil {
			return
		}

		policy = docker.RestartOnFailure(maxRetry)
		return
	}

	err = errors.New(fmt.Sprintf("Malformed restart policy %q", restart))
	return
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
	p1 := strings.SplitN(port, "@", 2)
	if len(p1) == 2 && d.env != nil && d.env.Name != p1[1] {
		return
	}

	p2 := strings.SplitN(p1[0], "/", 2)
	p3 := strings.SplitN(p2[0], ":", 3)

	if len(p2) != 2 || len(p3) != 3 {
		err = errors.New(fmt.Sprintf("Malformed port %q", port))
		return
	}

	guest = docker.Port(fmt.Sprintf("%s/%s", p3[2], p2[1]))
	host = docker.PortBinding{
		HostIP:   p3[0],
		HostPort: p3[1],
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

func (d *Docker) restartLinkedContainers(p *Project) error {
	failed := false
	for _, linked := range p.LinkedBy {
		list, err := d.ListContainers(linked)
		if err != nil {
			failed = true
			Error(err.Error(), "project", p)
			continue
		}

		for _, lc := range list {
			Info("Restarting linked container", "project", linked, "container", lc.GetShortId())
			if err := d.restartContainer(linked, lc); err != nil {
				failed = true
				Error("Unable to restart container", "project", linked, "container", lc.GetShortId())
			}
		}
	}

	if failed {
		return errors.New("Unable to restart one or more containers")
	}

	return nil
}

func (d *Docker) restartContainer(p *Project, c *Container) error {
	if !c.IsRunning() {
		return nil
	}

	if err := d.killContainer(c); err != nil {
		return err
	}

	if err := d.startContainer(p, c); err != nil {
		return err
	}

	return nil
}
