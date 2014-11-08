package core

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/sourcegraph/go-vcsurl"
)

const DEFAULT_BRANCH = "master"

type VCS string
type VCSInfo struct {
	Origin string
	Branch string
	*vcsurl.RepoInfo
}

func (v VCS) IsValid() bool {
	_, err := v.parse()
	return err == nil
}

func (v VCS) Info() *VCSInfo {
	info, _ := v.parse()
	return info
}

func (v VCS) parse() (*VCSInfo, error) {
	origin := string(v)
	branch := DEFAULT_BRANCH
	data := strings.SplitN(origin, "!", 2)
	if len(data) == 2 {
		branch = data[1]
	}

	info, err := vcsurl.Parse(data[0])
	return &VCSInfo{origin, branch, info}, err
}

type Commit string

type Revision map[VCS]Commit

func (r Revision) Get() string {
	if len(r) == 1 {
		for _, c := range r {
			return string(c)
		}
	}

	s := make([]string, 0)
	for _, commit := range r {
		s = append(s, string(commit))
	}

	sort.Strings(s)
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(s, ":"))))
}

func (r Revision) GetShort() string {
	id := r.Get()
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}

	return id[:shortLen]
}

func (r Revision) String() string {
	return r.Get()
}

type ImageId string

func (i ImageId) BelongsTo(p *Project) bool {
	return strings.HasPrefix(string(i), p.Name)
}

func (i ImageId) IsRevision(rev Revision) bool {
	s := strings.Split(string(i), ":")
	return strings.HasPrefix(s[1], rev.GetShort())
}

func (i ImageId) GetRevisionString() string {
	tmp := strings.SplitN(string(i), ":", 2)
	return tmp[1]
}

func (i ImageId) GetProjectString() string {
	tmp := strings.SplitN(string(i), ":", 2)
	return tmp[0]
}

type Image struct {
	DockerEndPoint string
	docker.APIImages
}

func (i Image) BelongsTo(p *Project) bool {
	for _, tag := range i.RepoTags {
		if strings.HasPrefix(tag, p.Name) {
			return true
		}
	}

	return false
}

func (i Image) GetRepoTagsAsImageId() []ImageId {
	r := make([]ImageId, 0)
	for _, tag := range i.RepoTags {
		r = append(r, ImageId(tag))
	}

	return r
}

type ImagesByCreated []*Image

func (c ImagesByCreated) Len() int           { return len(c) }
func (c ImagesByCreated) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ImagesByCreated) Less(i, j int) bool { return c[i].Created < c[j].Created }

var statusUp = regexp.MustCompile("^Up (.*)")

type Container struct {
	DockerEndPoint string
	Image          ImageId
	docker.APIContainers
}

func (c *Container) IsRunning() bool {
	return statusUp.MatchString(c.Status)
}

func (c *Container) GetPortsString() string {
	result := []string{}
	for _, port := range c.Ports {
		if port.IP == "" {
			result = append(result, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		} else {
			result = append(result, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
		}
	}
	return strings.Join(result, ", ")
}

func (c *Container) GetShortId() string {
	shortLen := 12
	if len(c.ID) < shortLen {
		shortLen = len(c.ID)
	}

	return c.ID[:shortLen]
}

func (c *Container) BelongsTo(p *Project) bool {
	if c.Image.BelongsTo(p) {
		return true
	}

	pname := fmt.Sprintf("/%s", p.Name)
	for _, name := range c.Names {
		if name == pname {
			return true
		}
	}

	return false
}

type Link struct {
	Project *Project
	Alias   string
}

func (l *Link) String() string {
	return fmt.Sprintf("%s:%s", l.Project.Name, l.Alias)
}

type LinkDefinition string

func (l LinkDefinition) GetProjectName() string {
	tmp := strings.SplitN(string(l), ":", 2)
	return tmp[0]
}

func (l LinkDefinition) GetAlias() string {
	tmp := strings.SplitN(string(l), ":", 2)
	return tmp[1]
}

type ContainersByCreated []*Container

func (c ContainersByCreated) Len() int           { return len(c) }
func (c ContainersByCreated) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ContainersByCreated) Less(i, j int) bool { return c[i].Created < c[j].Created }

type Environment struct {
	DockerEndPoints []string `gcfg:"DockerEndPoint"`
	EtcdServers     []string `gcfg:"EtcdServer"`
	Name            string
}

func (e *Environment) String() string {
	return e.Name
}

type Task int
type TaskStatus map[string]map[Task]time.Time

func (ts TaskStatus) Start(e *Environment, t Task) {
	if _, ok := ts[e.Name]; !ok {
		ts[e.Name] = make(map[Task]time.Time)
	}

	ts[e.Name][t] = time.Now()
}

func (ts TaskStatus) Stop(e *Environment, t Task) {
	if _, ok := ts[e.Name]; !ok {
		return
	}

	delete(ts[e.Name], t)

	if len(ts[e.Name]) == 0 {
		delete(ts, e.Name)
	}
}
