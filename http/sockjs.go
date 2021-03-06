package http

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mcuadros/dockership/core"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

type Message struct {
	Event   string
	Request map[string]string
}

type SockJSHandler func(msg Message, session sockjs.Session)

type SockJS struct {
	sessions []sockjs.Session
	handlers map[string]SockJSHandler
	sync.Mutex
}

func NewSockJS() *SockJS {
	return &SockJS{
		sessions: make([]sockjs.Session, 0),
		handlers: make(map[string]SockJSHandler, 0),
	}
}

func (s *SockJS) Send(event, data interface{}, isJSON bool) {
	var result []byte
	if isJSON {
		result = data.([]byte)
	} else {
		var err error
		result, err = json.Marshal(data)
		if err != nil {
			core.Error(fmt.Sprintf("Error SockJS send: %q", err.Error()))
			return
		}
	}

	raw := fmt.Sprintf("{\"event\":\"%s\", \"result\":%s}", event, result)

	for _, session := range s.sessions {
		session.Send(raw)
	}
}

func (s *SockJS) AddSessionAndRead(session sockjs.Session) {
	s.Lock()
	s.sessions = append(s.sessions, session)
	s.Unlock()

	s.onConnect(session)
	s.Read(session)
}

func (s *SockJS) onConnect(session sockjs.Session) {
	s.handleMessage(Message{Event: "connect"}, session)
}

func (s *SockJS) Read(session sockjs.Session) {
	for {
		if raw, err := session.Recv(); err == nil {
			s.processMessage(raw, session)
		} else {
			break
		}
	}
}

func (s *SockJS) processMessage(raw string, session sockjs.Session) {
	var msg Message
	if json.Unmarshal([]byte(raw), &msg) != nil {
		return
	}

	s.handleMessage(msg, session)
}

func (s *SockJS) handleMessage(msg Message, session sockjs.Session) {
	if h, ok := s.handlers[msg.Event]; ok {
		go h(msg, session)
	}
}

func (s *SockJS) AddHandler(event string, handler SockJSHandler) {
	s.handlers[event] = handler
}

type SockJSWriter struct {
	event     string
	sockjs    *SockJS
	formatter SockJSWriterFormatter
}

type SockJSWriterFormatter func(raw []byte) []byte

func NewSockJSWriter(sockjs *SockJS, event string) *SockJSWriter {
	return &SockJSWriter{
		event:  event,
		sockjs: sockjs,
		formatter: func(raw []byte) []byte {
			return raw
		},
	}
}

func (s *SockJSWriter) SetFormater(f SockJSWriterFormatter) {
	s.formatter = f
}

func (s *SockJSWriter) Write(raw []byte) (int, error) {
	s.sockjs.Send(s.event, s.formatter(raw), true)

	return len(raw), nil
}
