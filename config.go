package daemon

import (
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

var (
	registerConfig any = nil
	enableRemote   func(string, string) bool
)

func unmarshalConfig(c *mapstructure.DecoderConfig) { c.TagName = "json" }

func RegisterConfig(v any) { registerConfig = v }

// Config: Runtime > Remote > Local > Default

func readInConfig(project, version string) (err error) {
	os.MkdirAll("config", 0755)

	// Runtime
	if viper.ConfigFileUsed() != "" {
		if err = viper.ReadInConfig(); err != nil {
			return err
		}
		return nil
	}

	viper.SetConfigType("json")
	viper.AddConfigPath("config")
	viper.SetConfigName(InstanceTag)

	if enableRemote != nil && enableRemote(project, version) {
		return nil
	}

	// Local
	if err = viper.ReadInConfig(); err != nil {
		vlog.Warn("Failed to read local config", "err", err.Error())
	} else {
		vlog.Info("Load local config", "path", viper.ConfigFileUsed())
	}
	return nil
}
