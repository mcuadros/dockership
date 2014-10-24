package main

import (
	"net/http"

	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

	"github.com/gorilla/mux"
)

type StatusResult struct {
	Project *core.Project
	Status  map[string]*StatusRecord
	Error   []error
}

type StatusRecord struct {
	LastRevisionLabel string
	*core.ProjectStatus
}

func (s *server) HandleStatus(w http.ResponseWriter, r *http.Request) {
	Verbose()
	vars := mux.Vars(r)
	project := vars["project"]

	result := make(map[string]*StatusResult, 0)
	for name, p := range s.config.Projects {
		if project != "" && project != name {
			continue
		}

		record := &StatusResult{Project: p}
		sl, err := p.Status()
		if len(err) != 0 {
			record.Error = err
		} else {
			record.Status = make(map[string]*StatusRecord, 0)
			for _, s := range sl {
				record.Status[s.Enviroment.Name] = &StatusRecord{s.LastRevision.Get(), s}
			}
		}

		result[p.Name] = record
	}

	s.json(w, 200, result)
}
