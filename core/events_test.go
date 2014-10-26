package core

import (
	"sync"
	"sync/atomic"
)

import (
	. "gopkg.in/check.v1"
)

func (_ *CoreSuite) TestEventManager_Subscribe(c *C) {
	e := Event("foo")
	s := &Subscriber{}
	m := NewEventManager()

	c.Assert(m.Has(e, s), Equals, -1)
	m.Subscribe(e, s)
	c.Assert(m.Has(e, s), Equals, 0)
}

func (_ *CoreSuite) TestEventManager_Unsubscribe(c *C) {
	e := Event("foo")
	s := &Subscriber{}
	m := NewEventManager()

	m.Subscribe(e, s)
	m.Unsubscribe(e, s)
	c.Assert(m.Has(e, s), Equals, -1)
}

func (_ *CoreSuite) TestEventManager_Trigger(c *C) {
	var sync sync.WaitGroup
	var value int32 = 0

	e := Event("foo")
	s := &Subscriber{func(ctx ...interface{}) {
		defer sync.Done()
		atomic.StoreInt32(&value, ctx[0].(int32))
	}}

	m := NewEventManager()

	m.Subscribe(e, s)

	sync.Add(1)
	m.Trigger(e, int32(42))
	sync.Wait()

	c.Assert(value, Equals, int32(42))
}
