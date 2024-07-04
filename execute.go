package daemon

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

var (
	InstanceTag = "default"
)

type ActionFunc func(cmd *cobra.Command, args []string) error

func Execute(action ActionFunc) {
	rootCmd.RunE = action
	viper.BindPFlags(rootCmd.Flags())
	viper.SetEnvPrefix(rootCmd.Use)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := rootCmd.Execute(); err != nil {
		vlog.Errorf("%+v", err.Error())
	}
}
