package daemon

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type ActionFunc func(cmd *cobra.Command, args []string) error

var (
	std            *Daemon
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
		Use: "config", Aliases: []string{"c"}, Short: "Config",
	}
	configTemplateCmd = &cobra.Command{
		Use: "template json|yaml", Aliases: []string{"t"},
		Short: "Show Config Template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configData := viper.AllSettings()
			delete(configData, "config")
			delete(configData, "instance")
			var buf []byte
			switch args[0] {
			case "json":
				buf, _ = json.MarshalIndent(configData, "", "  ")
			case "yaml", "yml":
				buf, _ = yaml.Marshal(configData)
			}
			fmt.Println(string(buf))
			return nil
		},
	}
)

func init() {
	configCmd.AddCommand(configTemplateCmd)
	rootCmd.AddCommand(configCmd)
}

func (d *Daemon) RegisterConfig(config any) {
	registerConfig = config
}

func (d *Daemon) SetLogger(log *slog.Logger) {
	d.logger = log.WithGroup("daemon")
	d.systemd.logger = log.WithGroup("systemd")
	viper.WithLogger(log.WithGroup("viper"))
}

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }
func RootCmd() *cobra.Command           { return rootCmd }
func SetLogger(log *slog.Logger)        { std.SetLogger(log) }
func RegisterConfig(config any)         { std.RegisterConfig(config) }
func SetAction(action ActionFunc)       { rootCmd.RunE = action }

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
