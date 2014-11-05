package core

import (
	"bytes"
	"fmt"
)

type Dockerfile struct {
	blob        []byte
	project     *Project
	revision    Revision
	environment *Environment
}

func NewDockerfile(blob []byte, p *Project, r Revision, e *Environment) *Dockerfile {
	return &Dockerfile{
		blob:        blob,
		project:     p,
		revision:    r,
		environment: e,
	}
}

func (d *Dockerfile) Get() []byte {
	result := d.blob
	result = d.resolveInfoVariables(result)

	return result
}

func (d *Dockerfile) resolveInfoVariables(result []byte) []byte {
	if d.project == nil || d.environment == nil {
		return result
	}

	vars := map[string]string{
		"PROJECT": d.project.Name,
		"ENV":     d.environment.Name,
		"VCS":     string(d.project.Repository),
		"REV":     d.revision.GetShort(),
	}

	for name, value := range vars {
		varName := []byte(fmt.Sprintf("$DOCKERSHIP_%s", name))
		result = bytes.Replace(result, varName, []byte(value), -1)
	}

	return result
}
