package core

import (
	. "gopkg.in/check.v1"
)

func (_ *CoreSuite) TestLogger_Debug(c *C) {
	s := &Subscriber{}
	s.Handler = func(ctx ...interface{}) {
		Events.Unsubscribe(EventDebug, s)
		c.Assert(ctx[0], Equals, "foo")
	}

	Events.Subscribe(EventDebug, s)
	Debug("foo")
}

func (_ *CoreSuite) TestLogger_Info(c *C) {
	s := &Subscriber{}
	s.Handler = func(ctx ...interface{}) {
		Events.Unsubscribe(EventInfo, s)
		c.Assert(ctx[0], Equals, "foo")
	}

	Events.Subscribe(EventInfo, s)
	Info("foo")
}

func (_ *CoreSuite) TestLogger_Error(c *C) {
	s := &Subscriber{}
	s.Handler = func(ctx ...interface{}) {
		Events.Unsubscribe(EventError, s)
		c.Assert(ctx[0], Equals, "foo")
	}

	Events.Subscribe(EventError, s)
	Error("foo")
}

func (_ *CoreSuite) TestLogger_Warning(c *C) {
	s := &Subscriber{}
	s.Handler = func(ctx ...interface{}) {
		Events.Unsubscribe(EventWarning, s)
		c.Assert(ctx[0], Equals, "foo")
	}

	Events.Subscribe(EventWarning, s)
	Warning("foo")
}

func (_ *CoreSuite) TestLogger_Critical(c *C) {
	s := &Subscriber{}
	s.Handler = func(ctx ...interface{}) {
		Events.Unsubscribe(EventCritical, s)
		c.Assert(ctx[0], Equals, "foo")
	}

	Events.Subscribe(EventCritical, s)
	Critical("foo")
}
