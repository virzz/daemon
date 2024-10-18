package daemon

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

const defaultRemoteEndpoint = "config.app.virzz.com"

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

func ExecuteE(action ActionFunc) error {
	if std.logger == nil || std.systemd.logger == nil {
		std.SetLogger(vlog.Log)
	}
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) (err error) {
		instance, _ := cmd.PersistentFlags().GetString("instance")
		config, _ := cmd.PersistentFlags().GetString("config")
		viper.SetConfigType("json")
		if config != "" {
			viper.SetConfigFile(config)
		} else {
			viper.AddConfigPath(".")
			viper.SetConfigName("config_" + instance)
		}

		configLoaded := false
		if std.remoteConfig {
			remoteEndpoint, _ := cmd.PersistentFlags().GetString("remote-endpoint")
			if remoteEndpoint == "" {
				remoteEndpoint = std.remoteEndpoint
			}
			if remoteEndpoint == "" {
				remoteEndpoint = defaultRemoteEndpoint
			}
			key := fmt.Sprintf("/%s/%s/%s/%s", std.project, std.systemd.AppID, std.systemd.Version, instance)
			err = viper.AddSecureRemoteProvider("virzz", remoteEndpoint, key, string(std.secretKey))
			if err != nil {
				std.logger.Warn("Failed to add remote config provider", "err", err.Error())
			} else {
				err = viper.ReadRemoteConfig()
				if err != nil {
					std.logger.Warn("Failed to load remote config", "err", err.Error())
				} else {
					configLoaded = true
				}
			}
		}
		if !configLoaded {
			err = viper.ReadInConfig()
			if err != nil {
				return err
			}
		}
		return nil
	}
	rootCmd.RunE = action
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.BindPFlags(rootCmd.Flags())
	viper.SetEnvPrefix(rootCmd.Use)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
