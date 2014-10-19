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

func (s *CoreSuite) TestDocker_Deploy(c *C) {
	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)

	e := &Enviroment{DockerEndPoint: m.URL()}
	p := &Project{
		Repository: "git@github.com:foo/bar.git",
		Ports:      []string{"0.0.0.0:8080:80/tcp"},
	}

	rev := Revision{"foo": "bar"}
	err := NewDocker(e).Deploy(p, rev, []byte("FROM base\n"), false)
	c.Assert(err, Equals, nil)

	l, _ := NewDocker(e).ListContainers(p)
	c.Assert(l, HasLen, 1)
	c.Assert(l[0].Image.IsRevision(rev), Equals, true)
	c.Assert(l[0].IsRunning(), Equals, true)
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
	err := NewDocker(e).BuildImage(p, Revision{"key": "qux"}, []byte("FROM base\n"))
	s.Wait()

	c.Assert(err, Equals, nil)
	c.Assert(files, HasLen, 2)
	c.Assert(files["Dockerfile"], Equals, "FROM base\n")
	c.Assert(files[path.Base(file)], Equals, "qux")
	c.Assert(request.URL.Query().Get("t"), Equals, "foo/bar:qux")
	c.Assert(request.URL.Query().Get("nocache"), Equals, "1")
	c.Assert(request.URL.Query().Get("rm"), Equals, "1")
}

func (s *CoreSuite) TestDocker_Run(c *C) (p *Project, e *Enviroment, rev Revision) {
	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	d, _ := docker.NewClient(m.URL())

	e = &Enviroment{DockerEndPoint: m.URL()}
	p = &Project{Repository: "git@github.com:foo/bar.git", UseShortRevisions: true}

	buildImage(d, "foo/bar:qux")
	rev = Revision{"foo/bar": "qux"}
	err := NewDocker(e).Run(p, rev)
	c.Assert(err, Equals, nil)

	l, _ := NewDocker(e).ListContainers(p)
	c.Assert(l, HasLen, 1)
	c.Assert(l[0].Image.IsRevision(rev), Equals, true)
	c.Assert(l[0].IsRunning(), Equals, true)

	return
}

func (s *CoreSuite) TestDocker_Clean(c *C) {
	p, e, commit := s.TestDocker_Run(c)
	err := NewDocker(e).Clean(p, commit, false)
	c.Assert(err, Not(Equals), nil)
}

func (s *CoreSuite) TestDocker_CleanWithForce(c *C) {
	p, e, commit := s.TestDocker_Run(c)
	err := NewDocker(e).Clean(p, commit, true)
	c.Assert(err, Equals, nil)

	l, _ := NewDocker(e).ListContainers(p)
	c.Assert(l, HasLen, 0)
}

func (s *CoreSuite) TestDocker_formatPorts(c *C) {
	p := []string{
		"0.0.0.0:8080:80/tcp",
		"0.0.0.0:8080:80/udp",
		"0.0.0.0:42:42/tcp",
		"1.1.1.1:42:80/tcp",
	}

	r, _ := NewDocker(&Enviroment{}).formatPorts(p)
	c.Assert(r, HasLen, 3)
	c.Assert(r["80/tcp"], HasLen, 2)
	c.Assert(r["80/tcp"][0].HostIp, Equals, "0.0.0.0")
	c.Assert(r["80/tcp"][0].HostPort, Equals, "8080")
	c.Assert(r["80/udp"], HasLen, 1)
	c.Assert(r["42/tcp"], HasLen, 1)
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
