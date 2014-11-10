package http

import (
	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

func (s *server) EmitProjects(session sockjs.Session) {
	s.sockjs.Send("projects", s.config.Projects, false)
}

func (s *server) EmitUser(session sockjs.Session) {
	//user, _ := s.oauth.getUser(s.oauth.getToken(r))
	//s.sockjs.Send("user", user, false)
}
