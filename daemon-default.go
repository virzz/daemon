//go:build !remote
// +build !remote

package daemon

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

type Daemon struct {
	logger  *slog.Logger
	systemd *Systemd
}

func EnableRemoteConfig(project string, publicKey ...string) error {
	return std.EnableRemoteConfig(project, publicKey...)
}

func (d *Daemon) EnableRemoteConfig(project string, publicKey ...string) error {
	panic("remote config build with remote tag")
}

func (d *Daemon) ExecuteE(action ActionFunc) error {
	if std.logger == nil || std.systemd.logger == nil {
		std.SetLogger(vlog.Log)
	}
	if !slices.ContainsFunc(rootCmd.Commands(),
		func(cmd *cobra.Command) bool { return cmd.Use == "config" },
	) {
		rootCmd.AddCommand(configCmd)
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) (err error) {
		instance, _ := cmd.Flags().GetString("instance")
		config, _ := cmd.Flags().GetString("config")
		if config != "" {
			viper.SetConfigFile(config)
		} else {
			viper.SetConfigType("json")
			viper.AddConfigPath(".")
			viper.SetConfigName("config_" + instance)
		}
		err = viper.ReadInConfig()
		if err != nil {
			viper.SetConfigType("yaml")
			err = viper.ReadInConfig()
			if err != nil {
				vlog.Warn("Failed to read in config", "err", err.Error())
			}
		}
		if registerConfig != nil {
			err = viper.Unmarshal(registerConfig, func(dc *mapstructure.DecoderConfig) {
				dc.TagName = "json"
			})
			if err != nil {
				vlog.Error("Failed to unmarshal register config", "err", err.Error())
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

func ExecuteE(action ActionFunc) error {
	return std.ExecuteE(action)
}
