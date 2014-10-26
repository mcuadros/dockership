package core

var Events *EventManager

func init() {
	Events = NewEventManager()
}

var (
	EventDebug    Event = "log.debug"
	EventInfo     Event = "log.info"
	EventWarning  Event = "log.warning"
	EventError    Event = "log.error"
	EventCritical Event = "log.critical"
)

func Debug(ctx ...interface{}) {
	Events.Trigger(EventDebug, ctx...)
}

func Info(ctx ...interface{}) {
	Events.Trigger(EventInfo, ctx...)
}

func Warning(ctx ...interface{}) {
	Events.Trigger(EventWarning, ctx...)
}

func Error(ctx ...interface{}) {
	Events.Trigger(EventError, ctx...)
}

func Critical(ctx ...interface{}) {
	Events.Trigger(EventCritical, ctx...)
}
