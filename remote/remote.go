package remote

import (
	"io"

	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

type Client interface {
	Get() (io.Reader, error)
}

type Config struct {
	viper.RemoteProvider
}

func (c *Config) Get(rp viper.RemoteProvider) (io.Reader, error) {
	c.RemoteProvider = rp
	var err error
	if virzz == nil {
		virzz, err = NewVirzz(c)
		if err != nil {
			vlog.Error("Failed to create new client", "err", err.Error())
			return nil, err
		}
	}
	r, err := virzz.Get()
	if err != nil {
		vlog.Error("Failed to get remote config", "err", err.Error())
		return nil, err
	}
	return r, nil
}

func (c *Config) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	panic("unimplemented")
}
func (c *Config) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	panic("unimplemented")
}
