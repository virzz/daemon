package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	systemd "github.com/coreos/go-systemd/v22/dbus"
	"github.com/coreos/go-systemd/v22/unit"
	"github.com/virzz/vlog"
)

var InstanceTag = ""

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

func (d *Daemon) Version() string     { return d.version }
func (d *Daemon) Name() string        { return d.name }
func (d *Daemon) Description() string { return d.desc }

func (d *Daemon) Install(args ...string) error {
	vlog.Info("Install... " + d.name)
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	buf, err := CreateUnit(d.name, d.desc, execPath, args...)
	if err != nil {
		return err
	}
	os.WriteFile("/etc/systemd/system/"+d.name+"@.service", buf, 0644)
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	vlog.Info("Installed " + d.name)
	return conn.ReloadContext(ctx)
}

// Remove the service
func (d *Daemon) Remove() error {
	vlog.Info("Removing... " + d.name)
	err := d.Stop(true)
	if err != nil {
		vlog.Warn(err.Error())
	}
	err = os.Remove("/etc/systemd/system/" + d.name + "@.service")
	if err != nil {
		return err
	}
	vlog.Info("Removed " + d.name)
	return nil
}

// Start the service
func (d *Daemon) Start(num int, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	recv := make(chan string, 1)
	if num > 0 {
		for i := 1; i <= num; i++ {
			name := d.name + "@" + strconv.Itoa(i) + ".service"
			_, err = conn.StartUnitContext(ctx, name, "fail", recv)
			if err != nil {
				vlog.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				vlog.Error("Started [ " + name + " ] " + v)
			} else {
				vlog.Info("Started [ " + name + " ] " + v)
			}
		}
	} else if len(tags) > 0 {
		for _, tag := range tags {
			name := d.name + "@" + tag + ".service"
			_, err = conn.StartUnitContext(ctx, name, "fail", recv)
			if err != nil {
				vlog.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				vlog.Error("Started [ " + name + " ] " + v)
			} else {
				vlog.Info("Started [ " + name + " ] " + v)
			}
		}
	}
	return nil
}

// Stop the service
func (d *Daemon) Stop(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := d.Status(false)
		if err != nil {
			return err
		}
		recv := make(chan string, 1)
		for _, item := range items {
			_, err = conn.StopUnitContext(ctx, item.Name, "fail", recv)
			if err != nil {
				vlog.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				vlog.Error("Stop [ " + item.Name + "] " + v)
			} else {
				vlog.Info("Stop [ " + item.Name + " ] " + v)
			}
		}
	} else if len(tags) > 0 {
		recv := make(chan string, 1)
		for _, tag := range tags {
			name := d.name + "@" + tag + ".service"
			_, err = conn.StopUnitContext(ctx, name, "fail", recv)
			if err != nil {
				vlog.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				vlog.Error("Stop [" + name + "] " + v)
			} else {
				vlog.Info("Stop [ " + name + " ] " + v)
			}
		}
	}
	return nil
}

// Kill the service
func (d *Daemon) Kill(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := d.Status(false)
		if err != nil {
			return err
		}
		for _, item := range items {
			conn.KillUnitContext(ctx, item.Name, 9)
		}
	} else if len(tags) > 0 {
		for _, tag := range tags {
			conn.KillUnitContext(ctx, d.name+"@"+tag+".service", 9)
		}
	}
	return nil
}

// Restart the service
func (d *Daemon) Restart(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := d.Status(false)
		if err != nil {
			return err
		}
		recv := make(chan string, 1)
		for _, item := range items {
			_, err = conn.RestartUnitContext(ctx, item.Name, "fail", recv)
			if err != nil {
				vlog.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				vlog.Error("Restarted [ " + item.Name + "] " + v)
			} else {
				vlog.Info("Restarted [ " + item.Name + " ] " + v)
			}
		}
	} else if len(tags) > 0 {
		recv := make(chan string, 1)
		for _, tag := range tags {
			name := d.name + "@" + tag + ".service"
			_, err = conn.RestartUnitContext(ctx, name, "fail", recv)
			if err != nil {
				vlog.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				vlog.Error("Restarted [ " + name + " ] " + v)
			} else {
				vlog.Info("Restarted [ " + name + " ] " + v)
			}
		}
	}
	return nil
}

// Reload the service
func (d *Daemon) Reload(all bool, tags ...string) error {
	vlog.Info("Reloading... " + d.name)
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := d.Status(false)
		if err != nil {
			return err
		}
		recv := make(chan string, 1)
		for _, item := range items {
			_, err = conn.ReloadOrRestartUnitContext(ctx, item.Name, "fail", recv)
			if err != nil {
				return err
			}
			v := <-recv
			if v == "failed" {
				vlog.Error("Reloaded [ " + item.Name + "] " + v)
			} else {
				vlog.Info("Reloaded [ " + item.Name + " ] " + v)
			}
		}
	} else if len(tags) > 0 {
		recv := make(chan string, 1)
		for _, tag := range tags {
			name := d.name + "@" + tag + ".service"
			_, err = conn.ReloadOrRestartUnitContext(ctx, name, "fail", recv)
			if err != nil {
				vlog.Warn(err.Error())
			}
			v := <-recv
			if v == "failed" {
				vlog.Error("Reloaded [ " + name + " ] " + v)
			} else {
				vlog.Info("Reloaded [ " + name + " ] " + v)
			}
		}
	}
	return nil
}

// Status - Get service status
func (d *Daemon) Status(show bool) ([]systemd.UnitStatus, error) {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := conn.ListUnitsByPatternsContext(ctx, nil, []string{d.name + "*"})
	if err != nil {
		return nil, err
	}
	if show {
		for _, item := range items {
			if item.SubState == "running" {
				vlog.Info(item.Name, item.ActiveState, item.SubState)
			} else {
				vlog.Warn(item.Name, item.ActiveState, item.SubState)
			}
		}
	}
	return items, nil
}

func SetUnitConfig(section, name, value string) {
	if _, ok := unitConfig[section]; !ok {
		unitConfig[section] = make(map[string]string)
	}
	unitConfig[section][name] = value
}

var unitConfig = map[string]map[string]string{
	"Unit": {
		"Wants": "network.target",
	},
	"Service": {
		"Type":                     "exec",
		"ExecReload":               "/bin/kill -s HUP $MAINPID", // 发送HUP信号重载服务
		"Restart":                  "always",                    // 只要不是通过systemctl stop来停止服务，任何情况下都必须要重启服务
		"RestartSec":               "0",                         // 重启间隔
		"StartLimitInterval":       "30",                        // 启动尝试间隔
		"StartLimitBurst":          "10",                        // 最大启动尝试次数
		"RestartPreventExitStatus": "SIGKILL",                   // kill -9 不重启
	},
}

func CreateUnit(binName, desc, path string, args ...string) ([]byte, error) {
	binName += "@%i"
	if unitConfig == nil {
		return nil, fmt.Errorf("unitConfig is nil")
	}
	if _, ok := unitConfig["Unit"]; !ok {
		unitConfig["Unit"] = make(map[string]string)
	}
	if _, ok := unitConfig["Unit"]["Description"]; !ok {
		unitConfig["Unit"]["Description"] = strings.ToUpper(binName[:1]) + binName[1:] + " " + desc
	}
	if _, ok := unitConfig["Service"]; !ok {
		unitConfig["Service"] = make(map[string]string)
	}
	if _, ok := unitConfig["Service"]["WorkingDirectory"]; !ok {
		unitConfig["Service"]["WorkingDirectory"] = filepath.Dir(path)
	}
	if _, ok := unitConfig["Service"]["PIDFile"]; !ok {
		unitConfig["Service"]["PIDFile"] = "/run/" + binName + ".pid"
	}
	if _, ok := unitConfig["Service"]["ExecStartPre"]; !ok {
		unitConfig["Service"]["ExecStartPre"] = "/bin/rm -f /run/" + binName + ".pid"
	}
	if _, ok := unitConfig["Service"]["ExecStart"]; !ok {
		unitConfig["Service"]["ExecStart"] = path + " --daemon.instance %i " + strings.Join(args, " ")
	}
	if _, ok := unitConfig["Service"]["ExecStartPost"]; !ok {
		unitConfig["Service"]["ExecStartPost"] = "/bin/bash -c '/bin/systemctl show -p MainPID --value " + binName + " > /run/" + binName + ".pid'"
	}
	data := make([]*unit.UnitOption, 0, 10)
	for sec, v := range unitConfig {
		for name, value := range v {
			data = append(data, &unit.UnitOption{Section: sec, Name: name, Value: value})
		}
	}
	reader := unit.Serialize(data)
	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
