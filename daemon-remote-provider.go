//go:build remote
// +build remote

package daemon

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func padding(src []byte, blockSize int) []byte {
	p := blockSize - len(src)%blockSize
	return append(src, bytes.Repeat([]byte{byte(p)}, p)...)
}

func unpadding(src []byte) []byte {
	l := len(src)
	if n := int(src[l-1]); n <= l {
		return src[:l-n]
	}
	return src
}

// 解密
func decrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(buf, data)
	return unpadding(buf), nil
}

type RemoteProvider struct {
	viper.RemoteProvider
	EncryptSecret []byte
	TargetURL     string
	logger        *zap.Logger
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
		c.logger.Error("Failed to request remote", zap.Error(err))
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, rsp.Body)
		return nil, errors.Errorf("Failed to get remote config: %s", rsp.Status)
	}
	buf, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	if len(buf) < aes.BlockSize {
		return nil, errors.New("invalid remote config")
	}
	buf, err = decrypt(buf[aes.BlockSize:], []byte(rp.SecretKeyring()), buf[:aes.BlockSize])
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
