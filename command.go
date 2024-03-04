package daemon

import (
	"errors"
	"fmt"
	"os"
	"os/user"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/virzz/vlog"
)

type ActionFunc func(cmd *cobra.Command, args []string) error

var rootCmd = &cobra.Command{
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	SilenceErrors:     true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		InstanceTag = viper.GetString("daemon.instance")
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		panic("daemon action not implemented")
	},
}

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }
func RootCmd() *cobra.Command           { return rootCmd }

func wrapCmd(d *Daemon) *Daemon {
	rootCmd.Use = d.name
	rootCmd.Short = d.desc
	rootCmd.Version = d.version
	rootCmd.PersistentFlags().String("daemon.instance", "", "Get instance name from systemd template")
	// Daemon commands
	rootCmd.AddCommand(daemonCommand(d)...)
	return d
}

var persistentPreRunE = func(cmd *cobra.Command, args []string) error {
	_user, err := user.Current()
	if err != nil {
		return err
	}
	if _user.Gid == "0" || _user.Uid == "0" {
		return nil
	}
	return errors.New("root privileges required")
}

func daemonCommand(d *Daemon) []*cobra.Command {
	var installCmd = &cobra.Command{
		Use:               "install",
		Short:             "Install daemon",
		Aliases:           []string{"i"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(_ *cobra.Command, args []string) error {
			return d.Install(args...)
		},
	}
	var removeCmd = &cobra.Command{
		Use:               "remove",
		Short:             "Remove(Uninstall) daemon",
		Aliases:           []string{"rm", "uninstall", "uni"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(_ *cobra.Command, _ []string) error {
			return d.Remove()
		},
	}

	var startCmd = &cobra.Command{
		Use:               "start [tag]...",
		Short:             "Start daemon",
		Aliases:           []string{"run"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			num, _ := cmd.Flags().GetInt("num")
			return d.Start(num, args...)
		},
	}
	startCmd.Flags().IntP("num", "n", 0, "Num of Instances for start")

	var stopCmd = &cobra.Command{
		Use:               "stop",
		Short:             "Stop daemon",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return d.Stop(all, args...)
		},
	}
	stopCmd.Flags().BoolP("all", "a", false, "Stop all Instances")

	var restartCmd = &cobra.Command{
		Use:               "restart",
		Short:             "Restart daemon",
		Aliases:           []string{"r", "re"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return d.Restart(all, args...)
		},
	}
	restartCmd.Flags().BoolP("all", "a", false, "Restart all Instances")

	var killCmd = &cobra.Command{
		Use:               "kill",
		Short:             "Kill daemon",
		Aliases:           []string{"r", "re"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return d.Kill(all, args...)
		},
	}
	killCmd.Flags().BoolP("all", "a", false, "Kill all Instances")

	var reloadCmd = &cobra.Command{
		Use:               "reload",
		Short:             "Reload daemon",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return d.Reload(all, args...)
		},
	}
	reloadCmd.Flags().BoolP("all", "a", false, "Reload all Instances")

	var statusCmd = &cobra.Command{
		Use:               "status",
		Short:             "Status daemon",
		Aliases:           []string{"info", "if"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := d.Status(true)
			return err
		},
	}

	var versionCmd = &cobra.Command{
		Use:     "version",
		Short:   "Version daemon",
		Aliases: []string{"v"},
		Run: func(_ *cobra.Command, _ []string) {
			vlog.Info(d.Version())
		},
	}

	var systemdCmd = &cobra.Command{
		Hidden:            true,
		Use:               "unit",
		Short:             "print systemd unit service file",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			if t, _ := cmd.Flags().GetBool("template"); t {
				execPath, err := os.Executable()
				if err != nil {
					return err
				}
				buf, err := CreateUnit(d.name, d.desc, execPath, args...)
				if err != nil {
					return err
				}
				fmt.Println(string(buf))
				return nil
			}
			fn := "/etc/systemd/system/" + d.name + "@.service"
			vlog.Info("filepath = " + fn)
			buf, err := os.ReadFile(fn)
			if err != nil {
				return err
			}
			fmt.Println(string(buf))
			return nil
		},
	}
	systemdCmd.Flags().BoolP("template", "t", false, "Show template unit service file")

	return []*cobra.Command{installCmd, removeCmd, startCmd, stopCmd, reloadCmd,
		killCmd, restartCmd, statusCmd, versionCmd, systemdCmd}
}

func Execute(action ActionFunc) {
	viper.BindPFlags(rootCmd.PersistentFlags())
	viper.BindPFlags(rootCmd.Flags())
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/" + rootCmd.Use)
	viper.AddConfigPath("$HOME/.config/" + rootCmd.Use)
	viper.SetConfigName(rootCmd.Use)
	viper.SetConfigType("yaml")
	viper.ReadInConfig()
	viper.AutomaticEnv()
	rootCmd.RunE = action
	if err := rootCmd.Execute(); err != nil {
		vlog.Error(err.Error())
	}
}
