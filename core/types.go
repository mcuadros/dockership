package core

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"

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
	_, err := v.Info()
	return err == nil
}

func (v VCS) Info() (*VCSInfo, error) {
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

func (c Commit) GetShort() string {
	commit := string(c)
	shortLen := 12
	if len(commit) < shortLen {
		shortLen = len(commit)
	}

	return commit[:shortLen]
}

type Revision map[VCS]Commit

func (r Revision) Get() string {
	s := make([]string, 0)
	for _, commit := range r {
		s = append(s, string(commit))
	}

	sort.Strings(s)
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(s, ":"))))
}

func (r Revision) String() string {
	return r.Get()
}
