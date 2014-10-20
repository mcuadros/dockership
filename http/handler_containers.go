package main

import (
	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type ContainersRecord struct {
	Project   *core.Project
	Container *core.Container
	Error     string
}

func (s *server) HandleContainers(config config, params martini.Params, render render.Render) {
	Verbose()
	project := params["project"]

	r := make(map[string]*ContainersRecord, 0)
	for name, p := range config.Projects {
		if project != "" && project != name {
			continue
		}

		l, err := p.List()
		if err != nil {
			r[p.Name] = &ContainersRecord{Project: p, Error: err.Error()}
		} else {
			for _, c := range l {
				r[p.Name] = &ContainersRecord{Project: p, Container: c}
			}
		}
	}

	render.JSON(200, r)
}
