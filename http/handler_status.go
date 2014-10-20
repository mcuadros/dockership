package main

import (
	"fmt"

	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type StatusResult struct {
	Project *core.Project
	Status  map[string]*StatusRecord
	Error   string
}

type StatusRecord struct {
	LastRevisionLabel string
	*core.ProjectStatus
}

func (s *server) HandleStatus(config config, params martini.Params, render render.Render) {
	Verbose()
	project := params["project"]
	fmt.Println(project)

	r := make(map[string]*StatusResult, 0)
	for name, p := range config.Projects {
		if project != "" && project != name {
			continue
		}

		record := &StatusResult{Project: p}
		sl, err := p.Status()
		if err != nil {
			record.Error = err.Error()
		} else {
			record.Status = make(map[string]*StatusRecord, 0)
			for _, s := range sl {
				record.Status[s.Enviroment.Name] = &StatusRecord{s.LastRevision.Get(), s}
			}
		}

		r[p.Name] = record
	}

	render.JSON(200, r)
}
