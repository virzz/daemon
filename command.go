package daemon

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("name").Changed {
			std.name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flag("description").Changed {
			std.desc, _ = cmd.Flags().GetString("description")
		}
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		panic("daemon action not implemented")
	},
}

type ActionFunc func(cmd *cobra.Command, args []string) error

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }

func RootCmd() *cobra.Command { return rootCmd }

func Execute(action ...ActionFunc) {
	if len(action) > 0 && action[0] != nil {
		rootCmd.RunE = action[0]
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
	rootCmd.Use = d.Name()
	rootCmd.Short = d.Description()
	rootCmd.Version = v
	// Flags
	rootCmd.PersistentFlags().StringP("name", "n", "", "Modiry systemd service name")
	rootCmd.PersistentFlags().StringP("description", "d", "", "Modiry systemd service description")
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
		&cobra.Command{
			Hidden: true,
			Use:    "systemd", Short: "systemd service",
			RunE: func(_ *cobra.Command, args []string) error {
				buf, err := templateParseData("systemVConfig", d.GetTemplate(), templateData{
					Name:         d.name,
					Description:  d.desc,
					Author:       d.author,
					Dependencies: strings.Join(d.deps, " "),
					Args:         strings.Join(args, " "),
				})
				if err != nil {
					return err
				}
				fmt.Println("servicePath = ", d.servicePath(), "\n\n", string(buf))
				return nil
			},
		},
	)
	return d
}
