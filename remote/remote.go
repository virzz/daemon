package remote

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/viper"
	goetcdv3 "go.etcd.io/etcd/client/v3"
)

// Vipre Remote Config
//
// Provider: etcd3
// Path: /config/{programName}/{instance}

type Config struct {
	viper.RemoteProvider
	Username, Password string
}

func (c *Config) Get(rp viper.RemoteProvider) (io.Reader, error) {
	c.RemoteProvider = rp
	return c.get()
}

func (c *Config) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	c.RemoteProvider = rp
	return c.get()
}

func (c *Config) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	c.RemoteProvider = rp
	rr := make(chan *viper.RemoteResponse)
	stop := make(chan bool)
	go func() {
		for {
			client, err := c.newClient()
			if err != nil {
				time.Sleep(time.Duration(time.Second))
				continue
			}
			defer client.Close()
			ch := client.Watch(context.Background(), c.RemoteProvider.Path())
			select {
			case <-stop:
				return
			case res := <-ch:
				for _, event := range res.Events {
					rr <- &viper.RemoteResponse{Value: event.Kv.Value}
				}
			}
		}
	}()
	return rr, stop
}

func (c *Config) newClient() (*goetcdv3.Client, error) {
	return goetcdv3.New(goetcdv3.Config{
		Endpoints: []string{c.Endpoint()},
		Username:  c.Username, Password: c.Password,
	})
}

func (c *Config) get() (io.Reader, error) {
	client, err := c.newClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Get(ctx, c.Path())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("key not found")
	}
	return bytes.NewReader(resp.Kvs[0].Value), nil
}
