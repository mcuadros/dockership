package main

import (
	"github.com/mcuadros/dockership/core"

	"github.com/gin-gonic/gin"
)

func init() {
	router.GET("/containers/*project", func(ctx *gin.Context) {
		h, _ := NewHandlerContaienrs()
		h.Run(ctx)
	})
}

type HandlerContainers struct {
	config *core.Config
}

type ContainersRecord struct {
	Project   *core.Project
	Container *core.Container
	Error     string
}

func NewHandlerContaienrs() (*HandlerContainers, error) {
	var config core.Config
	config.LoadFile("config.ini")

	return &HandlerContainers{config: &config}, nil
}

func (h *HandlerContainers) Run(ctx *gin.Context) {
	project := ctx.Params.ByName("project")[1:]

	r := make(map[string]*ContainersRecord, 0)
	for name, p := range h.config.Projects {
		if project != "" && project != name {
			continue
		}

		l, err := p.List()
		if err != nil {
			r[p.Name] = &ContainersRecord{Project: p, Error: err.Error()}
		} else {
			for _, c := range l {
				r[p.Name] = &ContainersRecord{Project: p, Container: c}
			}
		}
	}

	ctx.JSON(200, r)
}
