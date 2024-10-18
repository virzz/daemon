//go:build !remote
// +build !remote

package daemon

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

func EnableRemoteConfig(project string, publicKey ...string) error {
	return std.EnableRemoteConfig(project, publicKey...)
}

func (d *Daemon) EnableRemoteConfig(project string, publicKey ...string) error {
	panic("remote config build with remote tag")
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
		return viper.ReadInConfig()
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
