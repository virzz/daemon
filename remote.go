//go:build remote
// +build remote

package daemon

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/virzz/vlog"

	"github.com/virzz/daemon/v2/remote"
)

func init() {
	enableRemote = loadRemoteConfig

	rootCmd.PersistentFlags().StringP("remote", "R", "http://config.virzz.com", "Remote config endpoint")
	rootCmd.PersistentFlags().StringP("secret", "S", "", "Remote config secret keyring")
	rootCmd.PersistentFlags().StringP("project", "P", "", "Remote config project")

	viper.RemoteConfig = &remote.Config{}
	viper.SupportedRemoteProviders = append(viper.SupportedRemoteProviders, "virzz")
}

func loadRemoteConfig(project, version string) bool {
	if project == "" {
		return false
	}
	viper.SetConfigType("json")
	key := fmt.Sprintf("/%s/%s/%s/%s", project, stdAppID, version, InstanceTag)
	viper.AddSecureRemoteProvider("virzz", viper.GetString("remote"), key, viper.GetString("secret"))
	err := viper.ReadRemoteConfig()
	if err != nil {
		vlog.Warn("Failed to read remote config", "key", key, "err", err.Error())
		return false
	}
	vlog.Info("Load Remote Config", "path", key)
	if err = viper.WriteConfig(); err != nil {
		vlog.Warn("can not save config", "err", err.Error())
	}
	var savePath = filepath.Join("config", InstanceTag+".json")
	if err = viper.WriteConfigAs(savePath); err != nil {
		vlog.Warn("can not save remote config", "err", err.Error())
	} else {
		vlog.Info("Save Remote Config to local", "path", savePath)
	}
	return true
}
