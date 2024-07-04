package daemon

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	SilenceErrors:     true,
	RunE: func(_ *cobra.Command, _ []string) error {
		panic("daemon action not implemented")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_ul := strings.Split(cmd.UseLine(), " ")
		if len(_ul) > 1 && slices.Contains([]string{"version", "completion", "help", "env"}, _ul[1]) {
			return nil
		}
		InstanceTag = viper.GetString("instance")
		return readInConfig(cmd.Root().Use)
	},
}

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }
func RootCmd() *cobra.Command           { return rootCmd }

func wrapCmd(d *Daemon) *Daemon {
	rootCmd.Use = d.name
	rootCmd.Short = d.desc
	rootCmd.Version = d.version

	rootCmd.PersistentFlags().StringP("instance", "i", "default", "Get instance name from systemd template")
	viper.BindPFlags(rootCmd.PersistentFlags())

	remoteFlagSet := pflag.NewFlagSet("remote", pflag.ContinueOnError)
	remoteFlagSet.Bool("remote.enable", false, "Enable Remote config")
	remoteFlagSet.String("config.type", "json", "Config type")
	remoteFlagSet.String("remote.endpoint", "", "Remote config endpoint")
	remoteFlagSet.String("remote.provider", "consul", "Remote config provider")
	remoteFlagSet.String("remote.username", "", "Remote config auth username")
	remoteFlagSet.String("remote.password", "", "Remote config auth password")
	remoteFlagSet.String("remote.project", "", "Remote config project name")
	remoteFlagSet.Bool("remote.save", true, "Remote config save to local")

	daemonViper.BindPFlags(remoteFlagSet)
	daemonViper.SetEnvPrefix("virzz_daemon")
	daemonViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	daemonViper.AutomaticEnv()

	rootCmd.PersistentFlags().AddFlagSet(remoteFlagSet)
	// Daemon commands
	rootCmd.AddCommand(daemonCommand(d)...)
	return d
}
