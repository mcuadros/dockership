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
	go func() {
		for {
			ctx.Writer.Flush()
			time.Sleep(1 * time.Millisecond)
		}
	}()

	Streaming(ctx.Writer)

	force := true
	project := ctx.Params.ByName("project")
	enviroment := ctx.Params.ByName("enviroment")

	if p, ok := h.config.Projects[project]; ok {
		Info("Starting deploy", "project", p, "enviroment", enviroment, "force", force)
		err := p.Deploy(force, enviroment)
		if err != nil {
			Critical(err.Error(), "project", project)

		}

		p.Test(enviroment)
		Info("Deploy success", "project", p, "enviroment", enviroment)

	} else {
		ctx.JSON(404, fmt.Sprintf("Project %q not found", project))
	}

}
