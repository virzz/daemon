package daemon

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/virzz/daemon/v2/remote"
	"github.com/virzz/vlog"
)

var daemonViper = viper.New()

var (
	registerConfig any = nil
)

func RegisterConfig(v any) { registerConfig = v }

func readInConfig(programName string) (err error) {
	os.MkdirAll("config", 0755)
	// Config: Runtime > Remote > Local
	// Runtime
	if viper.ConfigFileUsed() != "" {
		err = viper.ReadInConfig()
		if err != nil {
			return err
		}
		if registerConfig != nil {
			if err = viper.Unmarshal(registerConfig); err != nil {
				vlog.Error("Failed to unmarshal config", "err", err.Error())
				return
			}
		}
		return
	}
	configType := daemonViper.GetString("config.type")
	viper.SetConfigType(configType)
	// Remote
	endpoint := daemonViper.GetString("remote.endpoint")
	if endpoint != "" {
		daemonViper.SetConfigType(configType)
		projectName := daemonViper.GetString("remote.project")
		provider := daemonViper.GetString("remote.provider")
		remotePath := fmt.Sprintf("/config/%s/%s/%s", programName, projectName, InstanceTag)
		viper.RemoteConfig = &remote.Config{
			Username: daemonViper.GetString("remote.username"),
			Password: daemonViper.GetString("remote.password"),
		}
		err = viper.AddRemoteProvider(provider, endpoint, remotePath)
		if err != nil {
			vlog.Error("Failed to add remote provider", "err", err.Error())
		} else {
			if err = viper.ReadRemoteConfig(); err != nil {
				vlog.Warn("Failed to read remote config", "key", remotePath, "err", err.Error())
			} else {
				vlog.Info("Load Remote Config", "path", remotePath)
				if registerConfig != nil {
					if err = viper.Unmarshal(registerConfig); err != nil {
						vlog.Error("Failed to unmarshal config", "err", err.Error())
						return
					}
				}
				if daemonViper.GetBool("remote.save") {
					var savePath = filepath.Join("config", InstanceTag+"."+configType)
					os.MkdirAll(filepath.Dir(savePath), 0755)
					viper.SetConfigFile(savePath)
					if err = viper.WriteConfigAs(savePath); err != nil {
						vlog.Error("Failed to save remote config", "err", err.Error())
						return err
					}
					vlog.Info("Save Remote Config to local", "path", savePath)
				}
				return nil
			}
		}
	}
	// Local
	viper.SetConfigName(InstanceTag)
	viper.AddConfigPath("config")
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}
	if registerConfig != nil {
		if err = viper.Unmarshal(registerConfig); err != nil {
			vlog.Error("Failed to unmarshal config", "err", err.Error())
			return
		}
	}
	vlog.Info("Load local config", "path", viper.ConfigFileUsed())
	return nil
}
