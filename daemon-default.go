//go:build !remote
// +build !remote

package daemon

import (
	"slices"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Daemon struct {
	logger  *zap.Logger
	systemd *Systemd
}

func (d *Daemon) EnableRemote(project string, publicKey ...string) error {
	panic("remote config build with remote tag")
}

func (d *Daemon) ExecuteE(action ActionFunc) error {
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
				d.logger.Warn("Failed to read in config", zap.Error(err))
			}
		}
		if registerConfig != nil {
			err = viper.Unmarshal(registerConfig, func(dc *mapstructure.DecoderConfig) {
				dc.TagName = "json"
			})
			if err != nil {
				d.logger.Error("Failed to unmarshal register config", zap.Error(err))
				return err
			}
		}
		return nil
	}
	rootCmd.RunE = action
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.BindPFlags(rootCmd.Flags())
	viper.SetEnvPrefix(rootCmd.Use)
	viper.SetEnvKeyReplacer(strings.NewReplacer(
		".", "_",
		"-", "_",
	))
	viper.AutomaticEnv()
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
