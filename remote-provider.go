package daemon

import (
	"bytes"
	"crypto/aes"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/virzz/utils/crypto"
)

type RemoteProvider struct {
	viper.RemoteProvider
	EncryptSecret []byte
	TargetURL     string
	logger        *slog.Logger
}

func (c *RemoteProvider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	c.RemoteProvider = rp
	if c.TargetURL == "" {
		target, err := url.Parse(rp.Endpoint())
		if err != nil {
			return nil, err
		}
		if target.Host == "" {
			target.Host = defaultRemoteEndpoint
		}
		if target.Scheme == "" {
			target.Scheme = "https"
		}
		target.Path = rp.Path()
		c.TargetURL = target.String()
	}
	// Get remote config
	rsp, err := http.Post(c.TargetURL, "application/object-stream", bytes.NewBuffer(c.EncryptSecret))
	if err != nil {
		c.logger.Error("Failed to request remote", "err", err.Error())
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Failed to get remote config: %s", rsp.Status)
	}
	buf, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	if len(buf) < aes.BlockSize {
		return nil, errors.New("invalid remote config")
	}
	buf, err = crypto.AesDecrypt(buf[aes.BlockSize:], []byte(rp.SecretKeyring()), buf[:aes.BlockSize])
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

func (c *RemoteProvider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	panic("unimplemented")
}

func (c *RemoteProvider) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	panic("unimplemented")
}
