package main

import (
	"fmt"

	"github.com/mcuadros/dockership/core"
	. "github.com/mcuadros/dockership/logger"

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

type StatusResult struct {
	Project *core.Project
	Status  map[string]*StatusRecord
	Error   string
}

type StatusRecord struct {
	LastRevisionLabel string
	*core.ProjectStatus
}

func NewHandlerStatus() (*HandlerStatus, error) {
	var config core.Config
	if err := config.LoadFile("config.ini"); err != nil {
		panic(err)
	}

	for k, p := range config.Projects {
		fmt.Println("config", k, p.Name)
	}

	return &HandlerStatus{config: &config}, nil
}

func (h *HandlerStatus) Run(ctx *gin.Context) {
	Verbose()
	project := ctx.Params.ByName("project")[1:]

	r := make(map[string]*StatusResult, 0)
	for name, p := range h.config.Projects {
		if project != "" && project != name {
			continue
		}

		record := &StatusResult{Project: p}
		sl, err := p.Status()
		if err != nil {
			record.Error = err.Error()
		} else {
			record.Status = make(map[string]*StatusRecord, 0)
			for _, s := range sl {
				record.Status[s.Enviroment.Name] = &StatusRecord{s.LastRevision.Get(), s}
			}
		}

		r[p.Name] = record
	}

	ctx.JSON(200, r)
}
