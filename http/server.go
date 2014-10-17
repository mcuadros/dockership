package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var router = gin.Default()

func main() {
	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "static/index.html")
	})

	router.GET("/app.js", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "static/app.js")
	})

	router.Run(":8080")
}

type AutoFlusherWriter struct {
	writer     gin.ResponseWriter
	autoFlush  *time.Ticker
	closeChan  chan bool
	closedChan chan bool
}

func NewAutoFlusherWriter(writer gin.ResponseWriter, duration time.Duration) *AutoFlusherWriter {
	a := &AutoFlusherWriter{
		writer:     writer,
		autoFlush:  time.NewTicker(duration),
		closeChan:  make(chan bool),
		closedChan: make(chan bool),
	}
	go a.loop()

	return a
}

func (a *AutoFlusherWriter) loop() {
	for {
		select {
		case <-a.autoFlush.C:
			a.writer.Flush()
		case <-a.closeChan:
			a.writer.Flush()
			close(a.closedChan)
			return
		}
	}
}

func (a *AutoFlusherWriter) Write(p []byte) (int, error) {
	return a.writer.Write(p)
}

func (a *AutoFlusherWriter) Close() {
	for {
		select {
		case a.closeChan <- true:
		case <-a.closedChan:
			return
		}
	}
}
