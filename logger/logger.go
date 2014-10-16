package logger

import (
	"io"
	"os"

	"gopkg.in/inconshreveable/log15.v2"
)

var Log = log15.New()

func init() {
	Log.SetHandler(log15.LvlFilterHandler(log15.LvlInfo, log15.StdoutHandler))
}

func Verbose() {
	Log.SetHandler(log15.LvlFilterHandler(log15.LvlDebug, log15.StdoutHandler))
}

func Streaming(w io.Writer) {
	Log.SetHandler(
		log15.LvlFilterHandler(
			log15.LvlDebug,
			log15.StreamHandler(w, log15.JsonFormat()),
		),
	)
}

func Debug(msg string, ctx ...interface{}) {
	Log.Debug(msg, ctx...)
}

func Info(msg string, ctx ...interface{}) {
	Log.Info(msg, ctx...)
}

func Warning(msg string, ctx ...interface{}) {
	Log.Warn(msg, ctx...)
}

func Error(msg string, ctx ...interface{}) {
	Log.Error(msg, ctx...)
}

func Critical(msg string, ctx ...interface{}) {
	Log.Crit(msg, ctx...)
	os.Exit(1)
}
