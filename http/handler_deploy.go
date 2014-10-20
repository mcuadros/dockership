package main

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/mcuadros/dockership/logger"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
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
		_, err := p.Deploy(enviroment, force)
		if err != nil {
			Critical(err.Error(), "project", project)
		}

		time.Sleep(time.Second)
		Info("Deploy success", "project", p, "enviroment", enviroment)
	} else {
		render.JSON(404, fmt.Sprintf("Project %q not found", project))
	}
}
