package remote

import (
	"errors"
	"io"

	"github.com/spf13/viper"
)

type Config struct {
	viper.RemoteProvider
	Username, Password string
}

func (c *Config) Get(rp viper.RemoteProvider) (io.Reader, error) {
	c.RemoteProvider = rp
	switch rp.Provider() {
	case "etcdv3":
		return NewEtcdV3(c).Get()
	case "consul":
		return NewConsul(c).Get()
	}
	return nil, errors.New("Not suppost")
}

func (c *Config) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	panic("unimplemented")
}
func (c *Config) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	panic("unimplemented")
}
