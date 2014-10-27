package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mcuadros/dockership/core"

	"github.com/gorilla/mux"
)

func (s *server) HandleDeploy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writer := NewAutoFlusherWriter(w, 100*time.Millisecond)
	defer writer.Close()

	subs := subscribeWriteToEvents(writer)
	defer unsubscribeEvents(subs)

	force := true
	vars := mux.Vars(r)
	project := vars["project"]
	enviroment := vars["enviroment"]

	if p, ok := s.config.Projects[project]; ok {
		core.Info("Starting deploy", "project", p, "enviroment", enviroment, "force", force)
		err := p.Deploy(enviroment, force)
		if len(err) != 0 {
			for _, e := range err {
				core.Critical(e.Error(), "project", project)
			}
		} else {
			core.Info("Deploy success", "project", p, "enviroment", enviroment)
		}
	} else {
		s.json(w, 404, fmt.Sprintf("Project %q not found", project))
	}
}
