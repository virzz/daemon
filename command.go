package daemon

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/virzz/vlog"
)

var rootCmd = &cobra.Command{
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	SilenceErrors:     true,
	RunE: func(_ *cobra.Command, _ []string) error {
		panic("daemon action not implemented")
	},
}

func AddCommand(cmds ...*cobra.Command) { rootCmd.AddCommand(cmds...) }
func RootCmd() *cobra.Command           { return rootCmd }

var installCmd = &cobra.Command{
	GroupID:           "daemon",
	Use:               "install",
	Short:             "Install",
	Aliases:           []string{"i"},
	PersistentPreRunE: persistentPreRunE,
	RunE: func(_ *cobra.Command, args []string) error {
		return std.Install(args...)
	},
}

var removeCmd = &cobra.Command{
	GroupID:           "daemon",
	Use:               "remove",
	Short:             "Remove(Uninstall)",
	Aliases:           []string{"rm", "uninstall", "uni", "un"},
	PersistentPreRunE: persistentPreRunE,
	RunE: func(_ *cobra.Command, _ []string) error {
		return std.Remove()
	},
}
var startCmd = &cobra.Command{
	GroupID:           "daemon",
	Use:               "start [tag]...",
	Short:             "Start",
	Aliases:           []string{"run"},
	PersistentPreRunE: persistentPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		num, _ := cmd.Flags().GetInt("num")
		return std.Start(num, args...)
	},
}

var stopCmd = &cobra.Command{
	GroupID:           "daemon",
	Use:               "stop",
	Short:             "Stop",
	PersistentPreRunE: persistentPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		return std.Stop(all, args...)
	},
}

var restartCmd = &cobra.Command{
	GroupID:           "daemon",
	Use:               "restart",
	Short:             "Restart",
	Aliases:           []string{"r", "re"},
	PersistentPreRunE: persistentPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		return std.Restart(all, args...)
	},
}

var killCmd = &cobra.Command{
	GroupID:           "daemon",
	Use:               "kill",
	Short:             "Kill",
	Aliases:           []string{"k"},
	PersistentPreRunE: persistentPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		return std.Kill(all, args...)
	},
}

var reloadCmd = &cobra.Command{
	GroupID:           "daemon",
	Use:               "reload",
	Short:             "Reload",
	PersistentPreRunE: persistentPreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		return std.Reload(all, args...)
	},
}

var statusCmd = &cobra.Command{
	GroupID:           "daemon",
	Use:               "status",
	Short:             "Status",
	Aliases:           []string{"info", "if"},
	PersistentPreRunE: persistentPreRunE,
	RunE: func(_ *cobra.Command, _ []string) error {
		_, err := std.Status(true)
		return err
	},
}

func unitCmd(d *Daemon) *cobra.Command {
	var _unitCmd = &cobra.Command{
		GroupID:           "daemon",
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
	_unitCmd.Flags().BoolP("template", "t", false, "Show template unit service file")
	return _unitCmd
}
