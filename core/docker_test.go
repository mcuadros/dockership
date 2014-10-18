package core

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"

	"github.com/fsouza/go-dockerclient"
	"github.com/fsouza/go-dockerclient/testing"
	. "gopkg.in/check.v1"
)

func (s *CoreSuite) TestImageId_IsCommit(c *C) {
	i := ImageId("foo/bar:qux")

	c.Assert(i.IsCommit(Commit("qux")), Equals, true)
	c.Assert(i.IsCommit(Commit("bar")), Equals, false)

}

func (s *CoreSuite) TestImageId_BelongsTo(c *C) {
	i := ImageId("foo/bar:qux")

	c.Assert(i.BelongsTo(&Project{
		Repository: "git@github.com:foo/bar.git",
	}), Equals, true)

	c.Assert(i.BelongsTo(&Project{
		Repository: "git@github.com:qux/bar.git",
	}), Equals, false)
}

func (s *CoreSuite) TestImageId_GetCommit(c *C) {
	i := ImageId("foo/bar:qux")

	c.Assert(i.GetCommit(), Equals, Commit("qux"))
}

func (s *CoreSuite) TestDocker_BuildImage(c *C) {
	var request *http.Request
	files := make(map[string]string, 0)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		defer s.Done()

		tr := tar.NewReader(r.Body)
		for {
			header, err := tr.Next()
			if err != nil {
				break
			}

			content, _ := ioutil.ReadAll(tr)
			files[header.Name] = string(content)
		}

		request = r
	}))

	defer ts.Close()

	file := writeRandomFile("qux")
	e := &Enviroment{DockerEndPoint: ts.URL}
	p := &Project{
		Repository: "git@github.com:foo/bar.git",
		NoCache:    true,
		Files:      []string{file},
	}

	s.Add(1)
	err := NewDocker(e).BuildImage(p, Commit("foo"), []byte("FROM base\n"))
	s.Wait()

	c.Assert(err, Equals, nil)
	c.Assert(files, HasLen, 2)
	c.Assert(files["Dockerfile"], Equals, "FROM base\n")
	c.Assert(files[path.Base(file)], Equals, "qux")
	c.Assert(request.URL.Query().Get("t"), Equals, "foo/bar:foo")
	c.Assert(request.URL.Query().Get("nocache"), Equals, "1")
	c.Assert(request.URL.Query().Get("rm"), Equals, "1")
}

func (s *CoreSuite) TestDocker_Run(c *C) {
	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	d, _ := docker.NewClient(m.URL())

	e := &Enviroment{DockerEndPoint: m.URL()}
	p := &Project{Repository: "git@github.com:foo/bar.git", UseShortCommits: true}

	buildImage(d, "foo/bar:79bee4004ff1")
	commit := Commit("79bee4004ff184589afb4b547c77e88b")
	err := NewDocker(e).Run(p, commit)
	c.Assert(err, Equals, nil)

	l, _ := NewDocker(e).ListContainers(p)
	c.Assert(l, HasLen, 1)
	c.Assert(l[0].Image.IsCommit(commit), Equals, true)
	c.Assert(l[0].IsRunning(), Equals, true)
}

func buildImage(client *docker.Client, name string) {
	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	tr := tar.NewWriter(inputbuf)
	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: 10})
	tr.Write([]byte("FROM base\n"))
	tr.Close()

	opts := docker.BuildImageOptions{
		Name:         name,
		InputStream:  inputbuf,
		OutputStream: outputbuf,
	}

	if err := client.BuildImage(opts); err != nil {
		panic(err)
	}
}

func writeRandomFile(content string) string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}

	defer f.Close()
	if _, err := io.WriteString(f, content); err != nil {
		panic(err)
	}

	return f.Name()
}
