//go:build remote
// +build remote

package daemon

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const defaultPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAkhbTnK9POOy317u6ovxE
UqFT5FUPaTSSmAa0gDepG7B1SpDHmpsarJlf//doy9A4bqysxQ8Fu1njtxXU861s
J5lxS1p72UreuZoTbV+mnQFzeIqbPDiqQruzqws+hnKAVHdDcjy6NPvUH1na4bNf
snuVM/9FNik4bmd1bv362Oelhmj8jvx+sllf2L9/5H8/i35sW8oo811IE+cA+jow
BqvNT3/ayjtHlrYmnTOxGHv7H+j0JQ/yz2/ap7PWdIfspqGJZSV9iPKKfKfw37KF
H19ekMDgL248Y4PiK5BqWD4jY1hQfMsQf2ZVs2g6gGNQPAYLMiAXA4ngNurA4Kz+
xQIDAQAB
-----END PUBLIC KEY-----`

func EnableRemoteConfig(project string, publicKey ...string) error {
	return std.EnableRemoteConfig(project, publicKey...)
}

func (d *Daemon) EnableRemoteConfig(project string, publicKey ...string) error {
	rootCmd.PersistentFlags().String("remote-type", "json", "Remote config type")
	rootCmd.PersistentFlags().String("remote-endpoint", "", "Remote config endpoint")

	std.project = project
	std.remoteConfig = true
	std.secretKey = make([]byte, 32)
	io.ReadFull(rand.Reader, std.secretKey)

	var block *pem.Block
	if len(publicKey) > 0 && publicKey[0] != "" {
		block, _ = pem.Decode([]byte(publicKey[0]))
	} else {
		block, _ = pem.Decode([]byte(defaultPublicKey))
	}
	if block == nil || block.Type != "PUBLIC KEY" {
		return errors.New("Failed to decode PEM block containing public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	if key, ok := pub.(*rsa.PublicKey); ok {
		data, _ := rsa.EncryptOAEP(sha256.New(), rand.Reader, key, std.secretKey, nil)
		viper.RemoteConfig = &RemoteProvider{EncryptSecret: data, logger: std.logger.WithGroup("remote")}
		viper.SupportedRemoteProviders = append(viper.SupportedRemoteProviders, "virzz")
		return nil
	}
	return errors.New("not an RSA public key")
}
