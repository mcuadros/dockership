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
	environment := vars["environment"]

	if p, ok := s.config.Projects[project]; ok {
		core.Info("Starting deploy", "project", p, "environment", environment, "force", force)
		err := p.Deploy(environment, force)
		if len(err) != 0 {
			for _, e := range err {
				core.Critical(e.Error(), "project", project)
			}
		} else {
			core.Info("Deploy success", "project", p, "environment", environment)
		}
	} else {
		s.json(w, 404, fmt.Sprintf("Project %q not found", project))
	}
}
