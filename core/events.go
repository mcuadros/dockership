package core

import (
	"sync"
)

type EventManager struct {
	subs map[Event][]*Subscriber
	sync.Mutex
}

func NewEventManager() *EventManager {
	return &EventManager{
		subs: make(map[Event][]*Subscriber, 0),
	}
}

func (m *EventManager) Subscribe(e Event, s *Subscriber) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.subs[e]; !ok {
		m.subs[e] = make([]*Subscriber, 0)
	}

	m.subs[e] = append(m.subs[e], s)
}

func (m *EventManager) Unsubscribe(e Event, s *Subscriber) {
	m.Lock()
	defer m.Unlock()

	pos := m.Has(e, s)
	if pos < 0 {
		return
	}

	m.subs[e] = m.subs[e][:pos+copy(m.subs[e][pos:], m.subs[e][pos+1:])]
}

func (m *EventManager) Trigger(e Event, ctx ...interface{}) {
	m.Lock()
	defer m.Unlock()

	for _, sub := range m.subs[e] {
		if sub.Handler == nil {
			continue
		}

		go sub.Handler(ctx...)
	}
}

func (m *EventManager) Has(e Event, s *Subscriber) int {
	for i, sub := range m.subs[e] {
		if s == sub {
			return i
		}
	}

	return -1
}

type Subscriber struct {
	Handler func(...interface{})
}

type Event string
