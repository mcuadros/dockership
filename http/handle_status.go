package http

import (
	"github.com/mcuadros/dockership/core"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
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

func (s *server) HandleStatus(msg Message, session sockjs.Session) {
	var project string
	project, _ = msg.Request["project"]

	result := make(map[string]*StatusResult, 0)
	for name, p := range s.config.Projects {
		if project != "" && project != name {
			continue
		}

		record := &StatusResult{Project: p}
		sl, errs := p.Status()
		if len(errs) != 0 {
			for _, err := range errs {
				core.Error(err.Error(), "project", p)
			}

			record.Error = errs
		} else {
			record.Status = make(map[string]*StatusRecord, 0)
			for _, s := range sl {
				record.Status[s.Environment.Name] = &StatusRecord{s.LastRevision.Get(), s}
			}
		}

		result[p.Name] = record
	}

	s.sockjs.Send("status", result, false)
}
