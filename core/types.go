package core

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"sort"
	"strings"

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
	info := p.Repository.Info()
	return strings.HasPrefix(string(i), fmt.Sprintf("%s/%s", info.Username, info.Name))
}

func (i ImageId) IsRevision(rev Revision) bool {
	s := strings.Split(string(i), ":")
	return strings.HasPrefix(s[1], rev.GetShort())
}

func (i ImageId) GetRevisionString() string {
	tmp := strings.SplitN(string(i), ":", 2)
	return tmp[1]
}

var statusUp = regexp.MustCompile("^Up (.*)")

type Container struct {
	Enviroment *Enviroment
	Image      ImageId
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

type SortByCreated []*Container

func (c SortByCreated) Len() int           { return len(c) }
func (c SortByCreated) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c SortByCreated) Less(i, j int) bool { return c[i].Created < c[j].Created }

type Enviroment struct {
	DockerEndPoint string
	Name           string
	History        int `default:"3"`
}

func (e *Enviroment) String() string {
	return e.Name
}
