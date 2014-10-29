package core

import (
	"net/http"

	. "gopkg.in/check.v1"
)

func (_ *CoreSuite) TestEtcd_Get(c *C) {
	go startEtcdMockServer()

	e := NewEtcd([]string{"http://127.0.0.1:3000/"})
	r, err := e.Get("foo")
	c.Assert(err, Equals, nil)
	c.Assert(r, Equals, "this is awesome")
}

func (_ *CoreSuite) TestEtcd_GetDir(c *C) {
	go startEtcdMockServer()

	e := NewEtcd([]string{"http://127.0.0.1:3000/"})
	_, err := e.Get("dir")
	c.Assert(err, ErrorMatches, "Key \"dir\" is a directory")
}

type reponseHandler struct {
	data string
}

func (rh *reponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(rh.data))
}

func startEtcdMockServer() {
	mux := http.NewServeMux()

	mux.Handle("/v2/keys/dir", &reponseHandler{
		"{\"action\":\"get\",\"node\":{\"key\":\"/dir\",\"dir\":true,\"modifiedIndex\":5,\"createdIndex\":5}}",
	})

	mux.Handle("/v2/keys/foo", &reponseHandler{
		"{\"action\":\"get\",\"node\":{\"key\":\"/mykey\",\"value\":\"this is awesome\",\"modifiedIndex\":3,\"createdIndex\":3}}",
	})

	http.ListenAndServe(":3000", mux)
}
