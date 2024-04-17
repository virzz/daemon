package daemon

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	SilenceErrors:     true,
	RunE: func(_ *cobra.Command, _ []string) error {
		panic("daemon action not implemented")
	},
}

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }
func RootCmd() *cobra.Command           { return rootCmd }

func wrapCmd(d *Daemon) *Daemon {
	rootCmd.Use = d.name
	rootCmd.Short = d.desc
	rootCmd.Version = d.version
	rootCmd.PersistentFlags().StringP("instance", "i", "", "Get instance name from systemd template")
	// Daemon commands
	rootCmd.AddCommand(daemonCommand(d)...)
	return d
}
