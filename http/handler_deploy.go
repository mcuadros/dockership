package main

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/mcuadros/dockership/logger"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

func (s *server) HandleDeploy(
	config config,
	params martini.Params,
	render render.Render,
	writer http.ResponseWriter,
) {
	w := NewAutoFlusherWriter(writer, 100*time.Millisecond)
	defer w.Close()
	Streaming(w)

	force := true
	project := params["project"]
	enviroment := params["enviroment"]

	if p, ok := config.Projects[project]; ok {
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
		render.JSON(404, fmt.Sprintf("Project %q not found", project))
	}
}
