package daemon

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

var (
	InstanceTag = ""
)

type ActionFunc func(cmd *cobra.Command, args []string) error

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("instance", "default")
}

func Execute(action ActionFunc) {
	rootCmd.RunE = action
	viper.BindPFlags(rootCmd.Flags())
	if err := rootCmd.Execute(); err != nil {
		vlog.Error(err.Error())
	}
}
