package http

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/mcuadros/dockership/core"

	"github.com/gorilla/mux"
)

func (s *server) HandleDeploy(w http.ResponseWriter, r *http.Request) {
	//writer := NewSocketioWriter(s.socketio, "deploy", "foo")

	//subs := subscribeWriteToEvents(writer)
	//defer unsubscribeEvents(subs)

	writer := bytes.NewBuffer([]byte(""))
	force := true
	vars := mux.Vars(r)
	project := vars["project"]
	environment := vars["environment"]

	if p, ok := s.config.Projects[project]; ok {
		core.Info("Starting deploy", "project", p, "environment", environment, "force", force)
		err := p.Deploy(environment, writer, force)
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
