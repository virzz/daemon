package daemon

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

var (
	InstanceTag = "default"
	Project     = ""
)

type ActionFunc func(cmd *cobra.Command, args []string) error

// New - Create a new daemon
func New(appID, name, desc, version, commit string) (*Daemon, error) {
	std = &Daemon{name: strings.ToLower(name), desc: desc}
	stdAppID = appID

	rootCmd.Use = name
	rootCmd.Short = desc
	rootCmd.Version = stdAppID + " " + version + " " + commit

	rootCmd.PersistentFlags().StringP("instance", "i", "default", "Get instance name from systemd template")

	// Daemon commands
	rootCmd.AddGroup(&cobra.Group{ID: "daemon", Title: "Daemon commands"})
	rootCmd.AddCommand(
		installCmd, removeCmd, reloadCmd, unitCmd(std),
		startCmd, stopCmd, killCmd, restartCmd, statusCmd,
	)
	startCmd.Flags().IntP("num", "n", 0, "Num of Instances for start")
	stopCmd.Flags().BoolP("all", "a", false, "Stop all Instances")
	restartCmd.Flags().BoolP("all", "a", false, "Restart all Instances")
	killCmd.Flags().BoolP("all", "a", false, "Kill all Instances")
	reloadCmd.Flags().BoolP("all", "a", false, "Reload all Instances")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		InstanceTag = viper.GetString("instance")
		err := readInConfig(viper.GetString("project"), version)
		if err != nil {
			return err
		}
		if registerConfig != nil {
			if err := viper.Unmarshal(registerConfig, unmarshalConfig); err != nil {
				return err
			}
		}
		return nil
	}
	return std, nil
}

func Execute(action ActionFunc) {
	rootCmd.RunE = action
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.BindPFlags(rootCmd.Flags())
	viper.SetEnvPrefix(rootCmd.Use)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := rootCmd.Execute(); err != nil {
		vlog.Errorf("%+v", err.Error())
	}
}
