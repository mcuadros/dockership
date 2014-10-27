package main

import (
	"os"

	"github.com/mcuadros/dockership/core"

	"gopkg.in/inconshreveable/log15.v2"
)

func init() {
	log := log15.New()
	log.SetHandler(log15.LvlFilterHandler(log15.LvlInfo, log15.StdoutHandler))

	core.Events.Subscribe(core.EventInfo, &core.Subscriber{func(ctx ...interface{}) {
		log.Info(ctx[0].(string), ctx[1:]...)
	}})

	core.Events.Subscribe(core.EventDebug, &core.Subscriber{func(ctx ...interface{}) {
		log.Debug(ctx[0].(string), ctx[1:]...)
	}})

	core.Events.Subscribe(core.EventWarning, &core.Subscriber{func(ctx ...interface{}) {
		log.Warn(ctx[0].(string), ctx[1:]...)
	}})

	core.Events.Subscribe(core.EventError, &core.Subscriber{func(ctx ...interface{}) {
		log.Error(ctx[0].(string), ctx[1:]...)
	}})

	core.Events.Subscribe(core.EventCritical, &core.Subscriber{func(ctx ...interface{}) {
		log.Crit(ctx[0].(string), ctx[1:]...)
		os.Exit(1)
	}})
}
