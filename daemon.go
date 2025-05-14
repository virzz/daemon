package daemon

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	slogzap "github.com/samber/slog-zap/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type ActionFunc func(cmd *cobra.Command, args []string) error

var (
	std            *Daemon
	debug              = false
	registerConfig any = nil
	rootCmd            = &cobra.Command{
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
		SilenceErrors:     true,
		SilenceUsage:      true,
		RunE: func(_ *cobra.Command, _ []string) error {
			panic("daemon action not implemented")
		},
	}
	configCmd = &cobra.Command{
		Use: "config json|yaml", Aliases: []string{"c"},
		Short: "Show Config Template",
		RunE: func(cmd *cobra.Command, args []string) error {
			var buf []byte
			var config any
			if registerConfig != nil {
				config = registerConfig
			} else {
				config = viper.AllSettings()
				viper.Set("config", nil)
				viper.Set("instance", nil)
			}
			if len(args) > 0 && (args[0] == "yaml" || args[0] == "yml") {
				buf, _ = yaml.Marshal(config)
			} else {
				buf, _ = json.MarshalIndent(config, "", "  ")
			}
			fmt.Println(string(buf))
			return nil
		},
	}
)

func (d *Daemon) RegisterConfig(config any) { registerConfig = config }

func (d *Daemon) SetLogger(zlog *zap.Logger) {
	d.logger = zlog.Named("daemon")
	d.systemd.logger = d.logger.Named("systemd")
	if debug {
		viper.SetOptions(viper.WithLogger(
			slog.New(slogzap.Option{
				Level:  slog.LevelDebug,
				Logger: d.logger.Named("viper"),
			}.NewZapHandler()),
		))
		viper.Debug()
	}
}

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }
func RootCmd() *cobra.Command           { return rootCmd }
func SetLogger(log *zap.Logger)         { std.SetLogger(log) }
func RegisterConfig(config any)         { std.RegisterConfig(config) }
func SetAction(action ActionFunc)       { rootCmd.RunE = action }
func SetDebug()                         { debug = true }
func EnableRemote(project string, publicKey ...string) error {
	return std.EnableRemote(project, publicKey...)
}

func ExecuteE(action ActionFunc) error {
	if std.logger == nil || std.systemd.logger == nil {
		std.SetLogger(zap.L())
	}
	return std.ExecuteE(action)
}

func Execute(action ActionFunc) {
	if err := ExecuteE(action); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

// New - Create a new daemon
func New(appID, name, desc, version, commit string) *Daemon {
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
	return std
}
