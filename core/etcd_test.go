package core

import (
	"net/http"

	. "gopkg.in/check.v1"
)

func (_ *CoreSuite) TestEtcd_Get(c *C) {
	go startEtcdMockServer()

	e, err := NewEtcd([]string{"http://127.0.0.1:3000/"})
	c.Assert(err, IsNil)

	r, err := e.Get("foo")
	c.Assert(err, Equals, nil)
	c.Assert(r, Equals, "foofoo")
}

func (_ *CoreSuite) TestEtcd_GetDir(c *C) {
	go startEtcdMockServer()

	e, err := NewEtcd([]string{"http://127.0.0.1:3000/"})
	c.Assert(err, IsNil)

	_, err = e.Get("dir")
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
		"{\"action\":\"get\",\"node\":{\"key\":\"/mykey\",\"value\":\"foofoo\",\"modifiedIndex\":3,\"createdIndex\":3}}",
	})

	mux.Handle("/v2/keys/bar/foo", &reponseHandler{
		"{\"action\":\"get\",\"node\":{\"key\":\"/mykey\",\"value\":\"barfoobarfoo\",\"modifiedIndex\":3,\"createdIndex\":3}}",
	})

	http.ListenAndServe(":3000", mux)
}
