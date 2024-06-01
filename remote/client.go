package remote

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	goetcdv3 "go.etcd.io/etcd/client/v3"
)

type Client interface {
	Get() (io.Reader, error)
}

var _ Client = (*EtcdV3)(nil)
var _ Client = (*Consul)(nil)

type EtcdV3 struct{ cfg *Config }

func NewEtcdV3(cfg *Config) *EtcdV3 { return &EtcdV3{cfg: cfg} }

func (c *EtcdV3) newClient() (*goetcdv3.Client, error) {
	return goetcdv3.New(goetcdv3.Config{
		Endpoints: []string{c.cfg.Endpoint()},
		Username:  c.cfg.Username, Password: c.cfg.Password,
	})
}

func (c *EtcdV3) Get() (io.Reader, error) {
	client, err := c.newClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Get(ctx, c.cfg.Path())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("key not found")
	}
	return bytes.NewReader(resp.Kvs[0].Value), nil
}

type Consul struct{ cfg *Config }

func NewConsul(cfg *Config) *Consul { return &Consul{cfg: cfg} }

func (c *Consul) newClient() (*consul.KV, error) {
	conf := consul.DefaultConfig()
	conf.Address = c.cfg.Endpoint()
	conf.Token = c.cfg.Password
	client, err := consul.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return client.KV(), nil
}

func (c *Consul) Get() (io.Reader, error) {
	client, err := c.newClient()
	if err != nil {
		return nil, err
	}
	kv, _, err := client.Get(c.cfg.Path(), nil)
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, errors.New("Key was not found.")
	}
	return bytes.NewReader(kv.Value), nil
}
