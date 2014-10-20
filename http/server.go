package main

import (
	"net/http"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
)

var m = martini.Classic()

func main() {
	s := &server{}
	s.configure()
	s.run()
}

type server struct {
	config config
}

func (s *server) configure() {
	// status
	m.Get("/status", s.HandleStatus)
	m.Get("/status/:project", s.HandleStatus)

	// containers
	m.Get("/containers", s.HandleContainers)
	m.Get("/containers/:project", s.HandleContainers)

	// deploy
	m.Get("/deploy/:project/:enviroment", s.HandleDeploy)

	// assets
	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	m.Get("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/app.js")
	})

	// dic
	m.Use(render.Renderer(render.Options{}))
}

func (s *server) run() {
	if err := s.config.LoadFile("config.ini"); err != nil {
		panic(err)
	}

	m.Map(s.config)

	if err := http.ListenAndServe(s.config.HTTP.Listen, m); err != nil {
		panic(err)
	}
}

type AutoFlusherWriter struct {
	writer     http.ResponseWriter
	autoFlush  *time.Ticker
	closeChan  chan bool
	closedChan chan bool
}

func NewAutoFlusherWriter(writer http.ResponseWriter, duration time.Duration) *AutoFlusherWriter {
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
			a.writer.(http.Flusher).Flush()
		case <-a.closeChan:
			a.writer.(http.Flusher).Flush()
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
