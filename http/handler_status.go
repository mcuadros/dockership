package main

import (
	"fmt"
	"github.com/mcuadros/dockership/core"

	"github.com/gin-gonic/gin"
)

func init() {
	router.GET("/status/*project", func(ctx *gin.Context) {
		h, _ := NewHandlerStatus()
		h.Run(ctx)
	})
}

type HandlerStatus struct {
	config *core.Config
}

type StatusRecord struct {
	Project *core.Project
	Status  map[string]*core.ProjectStatus
	Error   string
}

func NewHandlerStatus() (*HandlerStatus, error) {
	var config core.Config
	config.LoadFile("config.ini")

	return &HandlerStatus{config: &config}, nil
}

func (c *HandlerStatus) Run(ctx *gin.Context) {
	project := ctx.Params.ByName("project")[1:]

	r := make(map[string]*StatusRecord, 0)
	for name, p := range c.config.Projects {
		if project != "" && project != name {
			continue
		}

		record := &StatusRecord{Project: p}
		sl, err := p.Status()
		if err != nil {
			record.Error = err.Error()
		} else {
			record.Status = make(map[string]*core.ProjectStatus, 0)
			for _, s := range sl {
				record.Status[s.Enviroment.Name] = s
			}
		}

		r[p.Name] = record
	}

	ctx.JSON(200, r)
}
