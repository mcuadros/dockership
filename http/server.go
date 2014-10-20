package main

import (
	"net/http"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/sessions"
)

func main() {
	s := &server{martini: martini.Classic()}
	s.readConfig()
	s.configure()
	s.configureAuth()
	s.run()
}

type server struct {
	martini *martini.ClassicMartini
	config  config
}

func (s *server) configure() {
	// status
	s.martini.Get("/status", s.HandleStatus)
	s.martini.Get("/status/:project", s.HandleStatus)

	// containers
	s.martini.Get("/containers", s.HandleContainers)
	s.martini.Get("/containers/:project", s.HandleContainers)

	// deploy
	s.martini.Get("/deploy/:project/:enviroment", s.HandleDeploy)

	// assets
	s.martini.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	s.martini.Get("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/app.js")
	})

	s.martini.Get("/logout", func(sess sessions.Session) string {
		sess.Clear()
		return "cleared!"
	})

	// dic
	s.martini.Use(render.Renderer(render.Options{}))
}

func (s *server) readConfig() {
	if err := s.config.LoadFile("config.ini"); err != nil {
		panic(err)
	}

	s.martini.Map(s.config)
}

func (s *server) run() {

	if err := http.ListenAndServe(s.config.HTTP.Listen, s.martini); err != nil {
		panic(err)
	}
}

// AutoFlusherWrite
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
