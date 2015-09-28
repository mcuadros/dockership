package core

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
)

type Etcd struct {
	endPoints []string
	client    etcd.Client
	kAPI      etcd.KeysAPI
}

func NewEtcd(endPoints []string) (*Etcd, error) {
	Debug("Connected to etcd", "endPoints", strings.Join(endPoints, ", "))

	client, err := etcd.New(etcd.Config{
		Endpoints: endPoints,
		Transport: etcd.DefaultTransport,
	})

	if err != nil {
		return nil, err
	}

	return &Etcd{
		endPoints: endPoints,
		client:    client,
		kAPI:      etcd.NewKeysAPI(client),
	}, nil
}

func (e *Etcd) Get(key string) (string, error) {
	r, err := e.kAPI.Get(context.Background(), key, nil)
	if err != nil {
		return "", fmt.Errorf("Error retrieving %q: %s", key, err)
	}

	if r.Node.Dir {
		return "", fmt.Errorf("Key %q is a directory", key)
	}

	return r.Node.Value, err
}
