package main

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/mcuadros/dockership/logger"

	"github.com/gorilla/mux"
)

func (s *server) HandleDeploy(w http.ResponseWriter, r *http.Request) {
	writer := NewAutoFlusherWriter(w, 100*time.Millisecond)
	defer writer.Close()
	Streaming(writer)

	force := true
	vars := mux.Vars(r)
	project := vars["project"]
	enviroment := vars["enviroment"]

	if p, ok := s.config.Projects[project]; ok {
		Info("Starting deploy", "project", p, "enviroment", enviroment, "force", force)
		err := p.Deploy(enviroment, force)
		if len(err) != 0 {
			for _, e := range err {
				Critical(e.Error(), "project", project)
			}
		} else {
			Info("Deploy success", "project", p, "enviroment", enviroment)
		}
	} else {
		s.json(w, 404, fmt.Sprintf("Project %q not found", project))
	}
}
