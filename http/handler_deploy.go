package main

import (
	"fmt"
	"time"

	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

	"github.com/gin-gonic/gin"
)

func init() {
	router.GET("/deploy/:project/:enviroment", func(ctx *gin.Context) {
		h, _ := NewHandlerDeploy()
		h.Run(ctx)
	})
}

type HandlerDeploy struct {
	config *core.Config
}

func NewHandlerDeploy() (*HandlerDeploy, error) {
	var config core.Config
	config.LoadFile("config.ini")

	return &HandlerDeploy{config: &config}, nil
}

func (h *HandlerDeploy) Run(ctx *gin.Context) {
	w := NewAutoFlusherWriter(ctx.Writer, 100*time.Millisecond)
	defer w.Close()

	Streaming(w)

	force := true
	project := ctx.Params.ByName("project")
	enviroment := ctx.Params.ByName("enviroment")

	if p, ok := h.config.Projects[project]; ok {
		Info("Starting deploy", "project", p, "enviroment", enviroment, "force", force)
		_, err := p.Deploy(enviroment, force)
		if err != nil {
			Critical(err.Error(), "project", project)
		}

		Info("Deploy success", "project", p, "enviroment", enviroment)

	} else {
		ctx.JSON(404, fmt.Sprintf("Project %q not found", project))
	}

}
