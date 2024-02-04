package daemon

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

type ActionFunc func() error

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }

func Execute(action ...ActionFunc) {
	if len(action) > 0 && action[0] != nil {
		rootCmd.RunE = func(*cobra.Command, []string) error {
			return action[0]()
		}
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printResult(msg string, err error) {
	if err != nil {
		fmt.Println(errors.WithMessage(err, msg))
	} else {
		fmt.Println(msg)
	}
}

func wrapCmd(d *Daemon) *Daemon {
	v, _ := d.Version()
	rootCmd = &cobra.Command{
		Use: d.Name(), Short: d.Description(), Version: v,
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		RunE: func(_ *cobra.Command, _ []string) error {
			panic("daemon action not implemented")
		},
	}
	// Daemon commands
	rootCmd.AddCommand(
		&cobra.Command{
			Use: "install", Short: "install daemon", Aliases: []string{"i"},
			Run: func(_ *cobra.Command, args []string) {
				printResult(d.Install(args...))
			},
		},
		&cobra.Command{
			Use: "remove", Short: "Remove daemon",
			Aliases: []string{"rm", "uninstall"},
			Run: func(_ *cobra.Command, _ []string) {
				printResult(d.Remove())
			},
		},
		&cobra.Command{
			Use: "start", Short: "Start daemon",
			Run: func(_ *cobra.Command, _ []string) {
				printResult(d.Start())
			},
		},
		&cobra.Command{
			Use: "stop", Short: "Stop daemon",
			Run: func(_ *cobra.Command, _ []string) {
				printResult(d.Stop())
			},
		},
		&cobra.Command{
			Use: "restart", Short: "Restart daemon", Aliases: []string{"r"},
			Run: func(_ *cobra.Command, _ []string) {
				printResult(d.Restart())
			},
		},
		&cobra.Command{
			Use: "status", Short: "Status daemon", Aliases: []string{"info", "if"},
			Run: func(_ *cobra.Command, _ []string) {
				printResult(d.Status())
			},
		},
		&cobra.Command{
			Use: "version", Short: "Version daemon", Aliases: []string{"v"},
			Run: func(_ *cobra.Command, _ []string) {
				printResult(d.Version())
			},
		},
	)
	return d
}
