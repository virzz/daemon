package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/virzz/vlog"

	"github.com/virzz/daemon/v2/remote"
)

// Vipre Remote Config
//
// Provider: etcd3
// Path: /config/{programName}/{projectName}/{instance}
// 		- projectName: server or project name
// 	eg: /config/worker/host1/default

var daemonViper = viper.New()

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
		programName := cmd.Root().Use
		// Config: Runtime > Remote > Local
		// Runtime
		if viper.ConfigFileUsed() != "" {
			return viper.ReadInConfig()
		}
		configType := daemonViper.GetString("config.type")
		viper.SetConfigType(configType)
		// Remote
		endpoint := daemonViper.GetString("remote.endpoint")
		if endpoint != "" {
			daemonViper.SetConfigType(configType)
			projectName := daemonViper.GetString("remote.project")
			remotePath := fmt.Sprintf("/config/%s/%s/%s", programName, projectName, InstanceTag)
			viper.RemoteConfig = &remote.Config{
				Username: daemonViper.GetString("remote.username"),
				Password: daemonViper.GetString("remote.password"),
			}
			err := viper.AddRemoteProvider("etcd3", endpoint, remotePath)
			if err != nil {
				vlog.Error("Failed to add remote provider", "err", err.Error())
			} else {
				if err = viper.ReadRemoteConfig(); err != nil {
					vlog.Error("Failed to read remote config", "key", remotePath, "err", err.Error())
				} else {
					vlog.Info("Load Remote Config", "path", remotePath)
					if daemonViper.GetBool("remote.save") {
						var savePath string
						if InstanceTag != "" && InstanceTag != "default" {
							savePath = filepath.Join("config", InstanceTag, programName+"."+configType)
						} else {
							savePath = filepath.Join("config", programName+"."+configType)
						}
						os.MkdirAll(filepath.Dir(savePath), 0755)
						viper.SetConfigFile(savePath)
						if err = viper.WriteConfigAs(savePath); err != nil {
							vlog.Error("Failed to save remote config", "err", err.Error())
							return err
						}
						vlog.Info("Save Remote Config to local", "path", savePath)
					}
					if daemonViper.GetBool("remote.watch") {
						err = viper.GetViper().WatchRemoteConfigOnChannel()
						if err != nil {
							vlog.Error("Failed to read remote config", "err", err.Error())
							return err
						}
						vlog.Info("Watch Remote Config")
					}
					return nil
				}
			}
		}
		// Local
		viper.SetConfigName(programName)
		if InstanceTag != "" && InstanceTag != "default" {
			viper.AddConfigPath(filepath.Join("config", InstanceTag))
		}
		viper.AddConfigPath("config")
		viper.AddConfigPath("$HOME/.config/" + programName)
		viper.AddConfigPath("/etc/" + programName)
		return viper.ReadInConfig()
	},
}

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }
func RootCmd() *cobra.Command           { return rootCmd }

func wrapCmd(d *Daemon) *Daemon {
	rootCmd.Use = d.name
	rootCmd.Short = d.desc
	rootCmd.Version = d.version

	rootCmd.PersistentFlags().StringP("instance", "i", "", "Get instance name from systemd template")
	viper.BindPFlags(rootCmd.PersistentFlags())

	remoteFlagSet := pflag.NewFlagSet("remote", pflag.ContinueOnError)
	remoteFlagSet.String("config.type", "json", "Config type")
	remoteFlagSet.StringP("remote.endpoint", "e", "", "Remote config endpoint")
	remoteFlagSet.StringP("remote.username", "u", "", "Remote config auth username")
	remoteFlagSet.StringP("remote.password", "p", "", "Remote config auth password")
	remoteFlagSet.StringP("remote.project", "n", "", "Remote config project name")
	remoteFlagSet.BoolP("remote.save", "s", false, "Remote config save to local")
	remoteFlagSet.BoolP("remote.watch", "w", true, "Remote config watch change")

	daemonViper.BindPFlags(remoteFlagSet)
	daemonViper.SetEnvPrefix("virzz_daemon")
	daemonViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	daemonViper.AutomaticEnv()

	rootCmd.PersistentFlags().AddFlagSet(remoteFlagSet)
	// Daemon commands
	rootCmd.AddCommand(daemonCommand(d)...)
	return d
}
