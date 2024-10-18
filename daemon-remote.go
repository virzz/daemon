//go:build remote
// +build remote

package daemon

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

const (
	defaultPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAkhbTnK9POOy317u6ovxE
UqFT5FUPaTSSmAa0gDepG7B1SpDHmpsarJlf//doy9A4bqysxQ8Fu1njtxXU861s
J5lxS1p72UreuZoTbV+mnQFzeIqbPDiqQruzqws+hnKAVHdDcjy6NPvUH1na4bNf
snuVM/9FNik4bmd1bv362Oelhmj8jvx+sllf2L9/5H8/i35sW8oo811IE+cA+jow
BqvNT3/ayjtHlrYmnTOxGHv7H+j0JQ/yz2/ap7PWdIfspqGJZSV9iPKKfKfw37KF
H19ekMDgL248Y4PiK5BqWD4jY1hQfMsQf2ZVs2g6gGNQPAYLMiAXA4ngNurA4Kz+
xQIDAQAB
-----END PUBLIC KEY-----`
	defaultRemoteEndpoint = "config.app.virzz.com"
)

type Daemon struct {
	logger         *slog.Logger
	systemd        *Systemd
	project        string
	remoteEndpoint string
	remoteConfig   bool
	secretKey      []byte
}

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

func ExecuteE(action ActionFunc) error {
	if std.logger == nil || std.systemd.logger == nil {
		std.SetLogger(vlog.Log)
	}
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) (err error) {
		instance, _ := cmd.PersistentFlags().GetString("instance")
		config, _ := cmd.PersistentFlags().GetString("config")
		viper.SetConfigType("json")
		if config != "" {
			viper.SetConfigFile(config)
		} else {
			viper.AddConfigPath(".")
			viper.SetConfigName("config_" + instance)
		}

		configLoaded := false
		if std.remoteConfig {
			remoteEndpoint, _ := cmd.PersistentFlags().GetString("remote-endpoint")
			if remoteEndpoint == "" {
				remoteEndpoint = std.remoteEndpoint
			}
			if remoteEndpoint == "" {
				remoteEndpoint = defaultRemoteEndpoint
			}
			key := fmt.Sprintf("/%s/%s/%s/%s", std.project, std.systemd.AppID, std.systemd.Version, instance)
			err = viper.AddSecureRemoteProvider("virzz", remoteEndpoint, key, string(std.secretKey))
			if err != nil {
				std.logger.Warn("Failed to add remote config provider", "err", err.Error())
			} else {
				err = viper.ReadRemoteConfig()
				if err != nil {
					std.logger.Warn("Failed to load remote config", "err", err.Error())
				} else {
					configLoaded = true
				}
			}
		}
		if !configLoaded {
			err = viper.ReadInConfig()
			if err != nil {
				return err
			}
		}
		return nil
	}
	rootCmd.RunE = action
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.BindPFlags(rootCmd.Flags())
	viper.SetEnvPrefix(rootCmd.Use)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
