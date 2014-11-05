package http

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/mcuadros/dockership/config"
	"github.com/mcuadros/dockership/core"

	"github.com/gorilla/mux"
)

var configFile string

func Start(version, build string) {
	core.Info("Starting HTTP daemon", "version", version, "build", build)
	flag.StringVar(&configFile, "config", config.DEFAULT_CONFIG, "config file")
	flag.Parse()

	s := &server{serverId: fmt.Sprintf("dockership %s, build %s", version, build)}
	s.readConfig(configFile)
	s.configure()
	s.configureAuth()
	s.run()
}

type server struct {
	serverId string
	mux      *mux.Router
	oauth    *OAuth
	config   config.Config
}

func (s *server) configure() {
	s.mux = mux.NewRouter()

	// status
	s.mux.Path("/rest/status").Methods("GET").HandlerFunc(s.HandleStatus)
	s.mux.Path("/rest/status/{project:.*}").Methods("GET").HandlerFunc(s.HandleStatus)

	// containers
	s.mux.Path("/rest/containers").Methods("GET").HandlerFunc(s.HandleContainers)
	s.mux.Path("/rest/containers/{project:.*}").Methods("GET").HandlerFunc(s.HandleContainers)

	// deploy
	s.mux.Path("/rest/deploy/{project:.*}/{environment:.*}").Methods("GET").HandlerFunc(s.HandleDeploy)

	// logged-user
	s.mux.Path("/rest/user").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := s.oauth.getUser(s.oauth.getToken(r))
		s.json(w, 200, user)
	})

	// assets
	s.mux.Path("/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		content, _ := Asset("static/index.html")
		w.Write(content)
	})

	s.mux.Path("/dockership.png").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		content, _ := Asset("static/dockership.png")
		w.Write(content)
	})

	s.mux.Path("/app.js").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		content, _ := Asset("static/app.js")
		w.Write(content)
	})
}

func (s *server) configureAuth() {
	s.oauth = NewOAuth(&s.config)
}

func (s *server) readConfig(configFile string) {
	if err := s.config.LoadFile(configFile); err != nil {
		panic(err)
	}
}

func (s *server) run() {
	core.Info("HTTP server running", "host:port", s.config.HTTP.Listen)
	if err := http.ListenAndServe(s.config.HTTP.Listen, s); err != nil {
		panic(err)
	}
}

func (s *server) json(w http.ResponseWriter, code int, response interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.oauth.Handler(w, r) {
		core.Debug("Handling request", "url", r.URL)
		w.Header().Set("Server", s.serverId)
		s.mux.ServeHTTP(w, r)
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
