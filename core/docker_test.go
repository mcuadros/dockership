package core

import (
	"archive/tar"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/fsouza/go-dockerclient/testing"
	. "gopkg.in/check.v1"
)

var slowFlag = flag.Bool("slow", false, "Skips Slow tests")

func (s *CoreSuite) TestDocker_Deploy(c *C) {
	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)

	p := &Project{
		Name:       "foo",
		Repository: "git@github.com:foo/bar.git",
		Ports:      []string{"0.0.0.0:8080:80/tcp"},
	}

	d, _ := NewDocker(m.URL())
	rev := Revision{"foo": "bar"}
	err := d.Deploy(p, rev, &Dockerfile{blob: []byte("FROM base\n")}, false)
	c.Assert(err, Equals, nil)

	l, _ := d.ListContainers(p)
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
	p := &Project{
		Name:       "image",
		Repository: "git@github.com:foo/bar.git",
		NoCache:    true,
		Files:      []string{file},
	}

	s.Add(1)
	d, _ := NewDocker(ts.URL)
	err := d.BuildImage(p, Revision{"key": "qux"}, &Dockerfile{blob: []byte("FROM base\n")})
	s.Wait()

	c.Assert(err, Equals, nil)
	c.Assert(files, HasLen, 2)
	c.Assert(files["Dockerfile"], Equals, "FROM base\n")
	c.Assert(files[path.Base(file)], Equals, "qux")
	c.Assert(request.URL.Query().Get("t"), Equals, "image:qux")
	c.Assert(request.URL.Query().Get("nocache"), Equals, "1")
	c.Assert(request.URL.Query().Get("rm"), Equals, "1")
}

func (s *CoreSuite) TestDocker_Run(c *C) (p *Project, m *testing.DockerServer, rev Revision) {
	m, _ = testing.NewServer("127.0.0.1:0", nil, nil)
	d, _ := docker.NewClient(m.URL())

	p = &Project{Name: "foo", Repository: "git@github.com:foo/bar.git", UseShortRevisions: true}

	buildImage(d, "foo:qux")
	rev = Revision{"foo/bar": "qux"}

	dc, err := NewDocker(m.URL())
	c.Assert(err, Equals, nil)

	err = dc.Run(p, rev)
	c.Assert(err, Equals, nil)

	l, err := dc.ListContainers(p)
	c.Assert(err, Equals, nil)
	c.Assert(l, HasLen, 1)
	c.Assert(l[0].Image.IsRevision(rev), Equals, true)
	c.Assert(l[0].IsRunning(), Equals, true)
	c.Assert(l[0].Names, HasLen, 1)
	c.Assert(l[0].Names[0], Equals, "/foo")
	return
}

func (s *CoreSuite) TestDocker_RunLinked(c *C) {
	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	d, _ := docker.NewClient(m.URL())

	linked := &Project{Name: "qux", Repository: "git@github.com:qux/bar.git"}
	project := &Project{Name: "foo", Repository: "git@github.com:foo/bar.git"}

	project.Links = map[string]*Link{"x": &Link{
		Alias:   "qux",
		Project: linked,
	}}

	linked.LinkedBy = []*Project{project}

	buildImage(d, "foo:qux")
	buildImage(d, "qux:qux")

	dc, err := NewDocker(m.URL())
	c.Assert(err, Equals, nil)

	err = dc.Run(project, Revision{"foo/bar": "qux"})
	c.Assert(err, Equals, nil)
	time.Sleep(500 * time.Millisecond)

	l, err := dc.ListContainers(project)
	c.Assert(err, Equals, nil)
	c.Assert(l[0].IsRunning(), Equals, true)
	prev := l[0].ID

	err = dc.Run(linked, Revision{"qux/bar": "qux"})
	c.Assert(err, Equals, nil)
	time.Sleep(100 * time.Millisecond)

	l, err = dc.ListContainers(linked)
	c.Assert(err, Equals, nil)
	c.Assert(l[0].IsRunning(), Equals, true)
	c.Assert(l[0].ID, Not(Equals), prev)

	return
}

func (s *CoreSuite) TestDocker_Clean(c *C) {
	if !*slowFlag {
		c.Skip("-slowFlag not provided")
	}

	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	d, _ := docker.NewClient(m.URL())
	p := &Project{Name: "foo", Repository: "git@github.com:foo/bar.git", UseShortRevisions: true}
	docker, _ := NewDocker(m.URL())

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		buildImage(d, fmt.Sprintf("foo:%d", i))
		docker.Run(p, Revision{"foo/bar": Commit(fmt.Sprintf("%d", i))})
	}

	lc, _ := docker.ListContainers(p)
	c.Assert(lc, HasLen, 5)
	c.Assert(lc[0].Image.GetRevisionString(), Equals, "0")
	c.Assert(lc[4].Image.GetRevisionString(), Equals, "4")
	c.Assert(lc[4].IsRunning(), Equals, true)

	li, _ := docker.ListImages(p)
	c.Assert(li, HasLen, 5)

	p.History = 3
	err := docker.Clean(p)
	c.Assert(err, Equals, nil)

	lc, _ = docker.ListContainers(p)
	c.Assert(lc, HasLen, 0)

	li, _ = docker.ListImages(p)
	c.Assert(li, HasLen, 3)

	c.Assert(li[0].GetRepoTagsAsImageId()[0].GetRevisionString(), Equals, "2")
	c.Assert(li[1].GetRepoTagsAsImageId()[0].GetRevisionString(), Equals, "3")
	c.Assert(li[2].GetRepoTagsAsImageId()[0].GetRevisionString(), Equals, "4")

	p.History = -10
	err = docker.Clean(p)
	c.Assert(err, Equals, nil)

	lc, _ = docker.ListContainers(p)
	c.Assert(lc, HasLen, 0)
	li, _ = docker.ListImages(p)
	c.Assert(li, HasLen, 0)
}

func (s *CoreSuite) TestDocker_ListImages(c *C) {
	m, _ := testing.NewServer("127.0.0.1:0", nil, nil)
	d, _ := docker.NewClient(m.URL())
	docker, _ := NewDocker(m.URL())
	p := &Project{Name: "foo", Repository: "git@github.com:foo/bar.git", UseShortRevisions: true}

	buildImage(d, "foo:qux")
	buildImage(d, "foo:baz")
	buildImage(d, "qux:baz")

	l, _ := docker.ListImages(p)
	c.Assert(l, HasLen, 2)
	c.Assert(l[0].DockerEndPoint, Equals, m.URL())
	c.Assert(l[1].DockerEndPoint, Equals, m.URL())
}

func (s *CoreSuite) TestDocker_formatPorts(c *C) {
	p := []string{
		"0.0.0.0:8080:80/tcp",
		"0.0.0.0:8080:80/udp",
		"0.0.0.0:42:42/tcp",
		"1.1.1.1:42:80/tcp",
	}

	d, _ := NewDocker("")
	r, _ := d.formatPorts(p)
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
