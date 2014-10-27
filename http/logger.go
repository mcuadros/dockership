package main

import (
	"io"

	"github.com/mcuadros/dockership/core"

	"gopkg.in/inconshreveable/log15.v2"
)

func init() {
	log := log15.New()
	log.SetHandler(log15.LvlFilterHandler(log15.LvlDebug, log15.StdoutHandler))
	subscribeEvents(log)
}

func subscribeWriteToEvents(w io.Writer) map[core.Event]*core.Subscriber {
	log := log15.New()
	log.SetHandler(
		log15.LvlFilterHandler(
			log15.LvlDebug,
			log15.StreamHandler(w, log15.JsonFormat()),
		),
	)

	return subscribeEvents(log)
}

func subscribeEvents(logger log15.Logger) map[core.Event]*core.Subscriber {
	subs := make(map[core.Event]*core.Subscriber, 0)
	subs[core.EventInfo] = &core.Subscriber{func(ctx ...interface{}) {
		logger.Info(ctx[0].(string), ctx[1:]...)
	}}

	subs[core.EventDebug] = &core.Subscriber{func(ctx ...interface{}) {
		logger.Debug(ctx[0].(string), ctx[1:]...)
	}}

	subs[core.EventWarning] = &core.Subscriber{func(ctx ...interface{}) {
		logger.Warn(ctx[0].(string), ctx[1:]...)
	}}

	subs[core.EventError] = &core.Subscriber{func(ctx ...interface{}) {
		logger.Error(ctx[0].(string), ctx[1:]...)
	}}

	subs[core.EventCritical] = &core.Subscriber{func(ctx ...interface{}) {
		logger.Crit(ctx[0].(string), ctx[1:]...)
	}}

	for e, s := range subs {
		core.Events.Subscribe(e, s)
	}

	return subs
}

func unsubscribeEvents(subs map[core.Event]*core.Subscriber) {
	for e, s := range subs {
		core.Events.Unsubscribe(e, s)
	}
}
