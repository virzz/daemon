package daemon

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	std     *Daemon
	rootCmd = &cobra.Command{
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		SilenceErrors:     true,
		SilenceUsage:      true,
		RunE: func(_ *cobra.Command, _ []string) error {
			panic("daemon action not implemented")
		},
	}
)

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }
func RootCmd() *cobra.Command           { return rootCmd }
func SetLogger(log *slog.Logger)        { std.SetLogger(log) }

type Daemon struct {
	logger         *slog.Logger
	systemd        *Systemd
	project        string
	remoteEndpoint string
	remoteConfig   bool
	secretKey      []byte
}

func (d *Daemon) SetLogger(log *slog.Logger) {
	d.logger = log.WithGroup("daemon")
	d.systemd.logger = log.WithGroup("systemd")
	viper.WithLogger(log.WithGroup("viper"))
}

// New - Create a new daemon
func New(appID, name, desc, version, commit string) (*Daemon, error) {
	rootCmd.Use = name
	rootCmd.Short = desc
	rootCmd.Version = appID + " " + version + " " + commit
	rootCmd.PersistentFlags().StringP("instance", "i", "default", "Get instance name from systemd template")
	rootCmd.PersistentFlags().StringP("config", "c", "", "Set custom config file")
	std = &Daemon{
		systemd: &Systemd{
			Name:        strings.ToLower(name),
			Description: desc,
			Version:     version,
			AppID:       appID,
		},
	}
	std.systemd.Command(rootCmd)
	return std, nil
}

type ActionFunc func(cmd *cobra.Command, args []string) error

func Execute(action ActionFunc) {
	if err := ExecuteE(action); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
