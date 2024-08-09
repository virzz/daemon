package remote

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/virzz/utils/crypto"
)

var _ Client = (*Virzz)(nil)

type Virzz struct {
	cfg *Config
	URL string
}

var virzz *Virzz

func NewVirzz(cfg *Config) (*Virzz, error) {
	target, err := url.Parse(cfg.Endpoint())
	if err != nil {
		return nil, err
	}
	target.Path = cfg.Path()
	return &Virzz{cfg: cfg, URL: target.String()}, nil
}

func (c *Virzz) Get() (io.Reader, error) {
	rsp, err := http.Get(c.URL)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get remote config: %s", rsp.Status)
	}
	buf, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	var dst = make([]byte, base64.StdEncoding.DecodedLen(len(buf)))
	n, err := base64.StdEncoding.Decode(dst, buf)
	if err != nil {
		return nil, err
	}
	buf = dst[:n]
	if len(buf) < aes.BlockSize {
		return nil, errors.New("invalid remote config")
	}
	buf, err = crypto.AesDecrypt(buf[aes.BlockSize:], []byte(c.cfg.SecretKeyring()), buf[:aes.BlockSize])
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}
