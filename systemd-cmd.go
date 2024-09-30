package daemon

import (
	"errors"
	"fmt"
	"os"
	"os/user"

	"github.com/spf13/cobra"
)

func (s *Systemd) Command(rootCmd *cobra.Command) {
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

	var installCmd = &cobra.Command{
		GroupID:           "daemon",
		Use:               "install",
		Short:             "Install",
		Aliases:           []string{"i"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			multi, _ := cmd.Flags().GetBool("multi")
			return std.systemd.Install(multi, args...)
		},
	}

	var removeCmd = &cobra.Command{
		GroupID:           "daemon",
		Use:               "remove",
		Short:             "Remove(Uninstall)",
		Aliases:           []string{"rm", "uninstall", "uni", "un"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(_ *cobra.Command, _ []string) error {
			return std.systemd.Remove()
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
			return std.systemd.Start(num, args...)
		},
	}

	var stopCmd = &cobra.Command{
		GroupID:           "daemon",
		Use:               "stop",
		Short:             "Stop",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return std.systemd.Stop(all, args...)
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
			return std.systemd.Restart(all, args...)
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
			return std.systemd.Kill(all, args...)
		},
	}

	var reloadCmd = &cobra.Command{
		GroupID:           "daemon",
		Use:               "reload",
		Short:             "Reload",
		PersistentPreRunE: persistentPreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, _ := cmd.Flags().GetBool("all")
			return std.systemd.Reload(all, args...)
		},
	}

	var statusCmd = &cobra.Command{
		GroupID:           "daemon",
		Use:               "status",
		Short:             "Status",
		Aliases:           []string{"info", "if"},
		PersistentPreRunE: persistentPreRunE,
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := std.systemd.Status(true)
			return err
		},
	}

	var unitCmd = &cobra.Command{
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
				multi, _ := cmd.Flags().GetBool("multi")
				buf, err := CreateUnit(multi, s.Name, s.Description, execPath, args...)
				if err != nil {
					return err
				}
				fmt.Println(string(buf))
				return nil
			}
			fn := "/etc/systemd/system/" + s.Name + "@.service"
			s.logger.Info("filepath = " + fn)
			buf, err := os.ReadFile(fn)
			if err != nil {
				return err
			}
			fmt.Println(string(buf))
			return nil
		},
	}

	// Daemon commands
	rootCmd.AddGroup(&cobra.Group{ID: "daemon", Title: "Systemd commands"})
	rootCmd.AddCommand(
		installCmd, removeCmd, reloadCmd, unitCmd,
		startCmd, stopCmd, killCmd, restartCmd, statusCmd,
	)
	installCmd.Flags().BoolP("multi", "m", false, "Use template unit service")
	startCmd.Flags().IntP("num", "n", 0, "Num of Instances for start")
	stopCmd.Flags().BoolP("all", "a", false, "Stop all Instances")
	restartCmd.Flags().BoolP("all", "a", false, "Restart all Instances")
	killCmd.Flags().BoolP("all", "a", false, "Kill all Instances")
	reloadCmd.Flags().BoolP("all", "a", false, "Reload all Instances")
	unitCmd.Flags().BoolP("template", "t", false, "Show template unit service file")
	unitCmd.Flags().BoolP("multi", "m", false, "Use template unit service")
}
