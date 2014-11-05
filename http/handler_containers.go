package http

import (
	"net/http"

	"github.com/mcuadros/dockership/core"

	"github.com/gorilla/mux"
)

type ContainersRecord struct {
	Project   *core.Project
	Container *core.Container
	Error     []error
}

func (s *server) HandleContainers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]

	result := make([]*ContainersRecord, 0)
	for name, p := range s.config.Projects {
		if project != "" && project != name {
			continue
		}

		l, err := p.ListContainers()
		if len(err) != 0 {
			result = append(result, &ContainersRecord{Project: p, Error: err})
		} else {
			for _, c := range l {
				result = append(result, &ContainersRecord{Project: p, Container: c})
			}
		}
	}

	s.json(w, 200, result)
}
