package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	systemd "github.com/coreos/go-systemd/v22/dbus"
	"github.com/coreos/go-systemd/v22/unit"
	"github.com/virzz/vlog"
)

var std *Daemon

// New - Create a new daemon
func New(name, desc, version string) (*Daemon, error) {
	std = &Daemon{name: strings.ToLower(name), desc: desc, version: version}
	return wrapCmd(std), nil
}

type Daemon struct {
	name    string
	desc    string
	version string
}

func (r *Daemon) Version() string     { return r.version }
func (r *Daemon) Name() string        { return r.name }
func (r *Daemon) Description() string { return r.desc }

func (r *Daemon) Install(args ...string) error {
	vlog.Log.Info("Install... " + r.name)
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	buf, err := createUnit(r.name, r.desc, execPath, args...)
	if err != nil {
		return err
	}
	os.WriteFile("/etc/systemd/system/"+r.name+"@.service", buf, 0644)
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	vlog.Log.Info("Installed " + r.name)
	return conn.ReloadContext(ctx)
}

// Remove the service
func (r *Daemon) Remove() error {
	vlog.Log.Info("Removing... " + r.name)
	err := r.Stop(true)
	if err != nil {
		vlog.Log.Warn(err.Error())
	}
	err = os.Remove("/etc/systemd/system/" + r.name + "@.service")
	if err != nil {
		return err
	}
	vlog.Log.Info("Removed " + r.name)
	return nil
}

// Start the service
func (r *Daemon) Start(num int, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if num > 0 {
		for i := 1; i <= num; i++ {
			_, err = conn.StartUnitContext(ctx, fmt.Sprintf("%s@%d.service", r.name, i), "fail", nil)
			if err != nil {
				return err
			}
		}
	} else if len(tags) > 0 {
		for _, tag := range tags {
			_, err = conn.StartUnitContext(ctx, r.name+"@"+tag+".service", "fail", nil)
			if err != nil {
				vlog.Log.Warn(err.Error())
			}
		}
	}
	return nil
}

// Stop the service
func (r *Daemon) Stop(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := r.Status(false)
		if err != nil {
			return err
		}
		for _, item := range items {
			_, err = conn.StopUnitContext(ctx, item.Name, "fail", nil)
			if err != nil {
				return err
			}
		}
	} else if len(tags) > 0 {
		for _, tag := range tags {
			_, err = conn.StopUnitContext(ctx, r.name+"@"+tag+".service", "fail", nil)
			if err != nil {
				vlog.Log.Warn(err.Error())
			}
		}
	}
	return nil
}

// Kill the service
func (r *Daemon) Kill(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := r.Status(false)
		if err != nil {
			return err
		}
		for _, item := range items {
			conn.KillUnitContext(ctx, item.Name, 9)
		}
	} else if len(tags) > 0 {
		for _, tag := range tags {
			conn.KillUnitContext(ctx, r.name+"@"+tag+".service", 9)
		}
	}
	return nil
}

// Start the service
func (r *Daemon) Restart(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := r.Status(false)
		if err != nil {
			return err
		}
		for _, item := range items {
			_, err = conn.RestartUnitContext(ctx, item.Name, "fail", nil)
			if err != nil {
				return err
			}
		}
	} else if len(tags) > 0 {
		for _, tag := range tags {
			_, err = conn.RestartUnitContext(ctx, r.name+"@"+tag+".service", "fail", nil)
			if err != nil {
				vlog.Log.Warn(err.Error())
			}
		}
	}
	return nil
}

// Status - Get service status
func (r *Daemon) Status(show bool) ([]systemd.UnitStatus, error) {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := conn.ListUnitsByPatternsContext(ctx, nil, []string{r.name + "*"})
	if err != nil {
		return nil, err
	}
	if show {
		for _, item := range items {
			if item.SubState == "running" {
				vlog.Log.Info(item.Name, item.ActiveState, item.SubState)
			} else {
				vlog.Log.Warn(item.Name, item.ActiveState, item.SubState)
			}
		}
	}
	return items, nil
}

func createUnit(binName, desc, path string, args ...string) ([]byte, error) {
	binName += "@%i"
	reader := unit.Serialize([]*unit.UnitOption{
		// [Unit]
		{Section: "Unit", Name: "Description",
			Value: strings.ToUpper(binName[:1]) + binName[1:] + " " + desc},
		{Section: "Unit", Name: "Wants", Value: "network.target"},
		// [Service]
		{Section: "Service", Name: "Type", Value: "exec"},
		{Section: "Service", Name: "WorkingDirectory", Value: filepath.Dir(path)},
		{Section: "Service", Name: "PIDFile", Value: "/run/" + binName + ".pid"},
		{Section: "Service", Name: "ExecStartPre", Value: "/bin/rm -f /run/" + binName + ".pid"},
		{Section: "Service", Name: "ExecStart", Value: path + " " + strings.Join(args, " ")},
		{Section: "Service", Name: "ExecStartPost", Value: "/bin/bash -c '/bin/systemctl show -p MainPID --value " + binName + " > /run/" + binName + ".pid'"},
		{Section: "Service", Name: "ExecReload", Value: "/bin/kill -s HUP $MAINPID"},
		// 只要不是通过systemctl stop来停止服务，任何情况下都必须要重启服务
		{Section: "Service", Name: "Restart", Value: "always"},
		// 重启间隔，比异常后等待10(s)再进行启动
		{Section: "Service", Name: "RestartSec", Value: "10"},
		// 设置为0表示不限次数重启
		{Section: "Service", Name: "StartLimitInterval", Value: "0"},
		// kill -9 不重启
		{Section: "Service", Name: "RestartPreventExitStatus", Value: "SIGKILL"},
	})
	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
