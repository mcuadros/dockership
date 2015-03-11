package http

import (
	"fmt"

	"github.com/mcuadros/dockership/core"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

type ContainersRecord struct {
	Project   *core.Project
	Container *core.Container
	Error     []error
}

func (s *server) HandleContainers(msg Message, session sockjs.Session) {
	fmt.Println()

	project, ok := msg.Request["project"]
	if !ok {
		return
	}

	s.sockjs.Send("containers", s.GetContainers(project), false)
}

func (s *server) GetContainers(project string) []*ContainersRecord {
	result := make([]*ContainersRecord, 0)

	for name, p := range s.config.Projects {
		if project != "" && project != name {
			continue
		}

		l, err := p.ListContainers()
		if len(err) != 0 {
			result = append(result, &ContainersRecord{Project: p, Error: err})
		} else {
			for _, c := range l {
				result = append(result, &ContainersRecord{Project: p, Container: c})
			}
		}
	}

	return result
}
