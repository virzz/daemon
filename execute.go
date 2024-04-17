package daemon

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

var InstanceTag = ""

type ActionFunc func(cmd *cobra.Command, args []string) error

func Execute(action ActionFunc) {
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.BindPFlags(rootCmd.Flags())
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		InstanceTag = viper.GetString("instance")
		if InstanceTag != "" {
			viper.AddConfigPath(InstanceTag)
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/" + rootCmd.Use)
		viper.AddConfigPath("/etc/" + rootCmd.Use)
		viper.SetConfigName(rootCmd.Use)
		viper.SetConfigType("yaml")
		viper.ReadInConfig()
		viper.AutomaticEnv()
	}
	rootCmd.RunE = action
	if err := rootCmd.Execute(); err != nil {
		vlog.Error(err.Error())
	}
}
