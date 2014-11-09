package http

import (
	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

func (s *server) HandleConnect(msg Message, session sockjs.Session) {
	s.EmitProjects(session)
	s.EmitUser(session)
}
