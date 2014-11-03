package core

import (
	"sync"
)

type DockerGroup struct {
	environment *Environment
	dockers     map[string]*Docker
	sync.WaitGroup
}

func NewDockerGroup(environment *Environment) (*DockerGroup, error) {
	dg := &DockerGroup{
		environment: environment,
		dockers:     make(map[string]*Docker, 0),
	}

	for _, endPoint := range environment.DockerEndPoints {
		if d, err := NewDocker(endPoint); err == nil {
			dg.dockers[endPoint] = d
		} else {
			return nil, err
		}
	}

	return dg, nil
}

func (d *DockerGroup) Deploy(p *Project, rev Revision, dockerfile []byte, force bool) []error {
	Info("Deploying dockerfile", "project", p, "revision", rev, "end-points", len(d.dockers))
	return d.batchErrorResult(func(docker *Docker) interface{} {
		return &errorResult{err: docker.Deploy(p, rev, dockerfile, force)}
	})
}

func (d *DockerGroup) Clean(p *Project) []error {
	Info("Cleaning containers", "project", p, "end-points", len(d.dockers))
	return d.batchErrorResult(func(docker *Docker) interface{} {
		return &errorResult{err: docker.Clean(p)}
	})
}

type listContainersResult struct {
	containers []*Container
	err        error
}

func (d *DockerGroup) ListContainers(p *Project) ([]*Container, []error) {
	f := func(docker *Docker) interface{} {
		c, e := docker.ListContainers(p)
		return &listContainersResult{c, e}
	}

	errors := make([]error, 0)
	containers := make([]*Container, 0)
	for _, e := range d.batchInterfaceResult(f) {
		l := e.(*listContainersResult)
		if l.err != nil {
			errors = append(errors, l.err)
		}
		containers = append(containers, l.containers...)
	}

	return containers, errors
}

type listImagesResult struct {
	images []*Image
	err    error
}

func (d *DockerGroup) ListImages(p *Project) ([]*Image, []error) {
	f := func(docker *Docker) interface{} {
		c, e := docker.ListImages(p)
		return &listImagesResult{c, e}
	}

	errors := make([]error, 0)
	images := make([]*Image, 0)
	for _, e := range d.batchInterfaceResult(f) {
		l := e.(*listImagesResult)
		if l.err != nil {
			errors = append(errors, l.err)
		}
		images = append(images, l.images...)
	}

	return images, errors
}

func (d *DockerGroup) BuildImage(p *Project, rev Revision, dockerfile []byte) []error {
	Info("Building image", "project", p, "revision", rev, "end-points", len(d.dockers))
	return d.batchErrorResult(func(docker *Docker) interface{} {
		return &errorResult{err: docker.BuildImage(p, rev, dockerfile)}
	})
}

func (d *DockerGroup) Run(p *Project, rev Revision) []error {
	return d.batchErrorResult(func(docker *Docker) interface{} {
		return &errorResult{err: docker.Run(p, rev)}
	})
}

type errorResult struct{ err error }

func (d *DockerGroup) batchErrorResult(f func(docker *Docker) interface{}) []error {
	r := make([]error, 0)
	for _, e := range d.batchInterfaceResult(f) {
		if err := e.(*errorResult).err; err != nil {
			r = append(r, err)
		}
	}

	return r
}

func (d *DockerGroup) batchInterfaceResult(f func(docker *Docker) interface{}) []interface{} {
	count := len(d.dockers)
	c := make(chan interface{}, count)
	defer close(c)

	for _, docker := range d.dockers {
		d.Add(1)
		go func(docker *Docker) {
			defer d.Done()
			c <- f(docker)
		}(docker)
	}
	d.Wait()

	r := make([]interface{}, 0)
	for i := 0; i < count; i++ {
		r = append(r, <-c)
	}

	return r
}
