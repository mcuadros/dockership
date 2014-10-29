package core

import (
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

type Etcd struct {
	machines []string
	client   *etcd.Client
}

func NewEtcd(machines []string) *Etcd {
	Debug("Connected to etcd", "machines", strings.Join(machines, ", "))

	return &Etcd{client: etcd.NewClient(machines), machines: machines}
}

func (e *Etcd) Get(key string) (string, error) {
	r, err := e.client.Get(key, false, false)
	if err != nil {
		return "", err
	}

	if r.Node.Dir {
		return "", errors.New(fmt.Sprintf("Key %q is a directory", key))
	}

	return r.Node.Value, err
}
