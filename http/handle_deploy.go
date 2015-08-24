package http

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/mcuadros/dockership/core"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

var ErrProjectNotFound = errors.New("Project not found")

type DeployResult struct {
	Done    bool
	Elapsed time.Duration
	Errors  []error `json:",omitempty"`
}

func (s *server) HandleDeploy(msg Message, session sockjs.Session) {
	force := true
	project, ok := msg.Request["project"]
	if !ok {
		core.Error("Missing project", "request", "deploy")
		return
	}

	environment, ok := msg.Request["environment"]
	if !ok {
		core.Error("Missing environment", "request", "deploy")
		return
	}

	now := time.Now()

	writer := NewSockJSWriter(s.sockjs, "deploy")
	writer.SetFormater(func(raw []byte) []byte {
		str, _ := json.Marshal(map[string]string{
			"environment": environment,
			"project":     project,
			"date":        now.String(),
			"log":         string(raw),
		})

		return str
	})

	go func(session sockjs.Session) {
		time.Sleep(50 * time.Millisecond)
		s.EmitProjects(session)
	}(session)

	s.DoDeploy(writer, project, environment, force)
	s.EmitProjects(session)
}

func (s *server) DoDeploy(w io.Writer, project, environment string, force bool) *DeployResult {
	start := time.Now()
	r := &DeployResult{}
	defer func() {
		r.Elapsed = time.Since(start)
	}()

	core.Info(
		"Starting deploy",
		"project", project, "environment", environment, "force", force,
	)

	p, ok := s.config.Projects[project]
	if !ok {
		core.Error("Project not found", "project", p)

		r.Errors = []error{ErrProjectNotFound}
		return r
	}

	r.Errors = p.Deploy(environment, w, force)
	if len(r.Errors) == 0 {
		r.Done = true
		core.Info("Deploy success", "project", p, "environment", environment)
	} else {
		for _, e := range r.Errors {
			core.Critical(e.Error(), "project", p, "environment", environment)
		}
	}

	return r
}
