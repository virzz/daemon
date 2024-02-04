package daemon

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// New - Create a new daemon
func New(name, desc, version, author string, deps ...string) (*Daemon, error) {
	d := &Daemon{name: name, desc: desc, version: version, author: author, deps: deps}
	return wrapCmd(d), nil
}

type Daemon struct {
	name     string
	desc     string
	version  string
	author   string
	deps     []string
	template string
}

func (r *Daemon) Name() string             { return r.name }
func (r *Daemon) Description() string      { return r.desc }
func (r *Daemon) Author() string           { return r.author }
func (r *Daemon) Version() (string, error) { return r.version, nil }

func (linux *Daemon) servicePath() string {
	return "/etc/systemd/system/" + linux.name + ".service"
}

func (linux *Daemon) isInstalled() bool {
	_, err := os.Stat(linux.servicePath())
	return err == nil
}

func (linux *Daemon) checkRunning() (string, bool) {
	output, err := exec.Command("systemctl", "status", linux.name+".service").Output()
	if err == nil {
		if matched, err := regexp.MatchString("Active: active", string(output)); err == nil && matched {
			reg := regexp.MustCompile("Main PID: ([0-9]+)")
			data := reg.FindStringSubmatch(string(output))
			if len(data) > 1 {
				return "Service (pid  " + data[1] + ") is running...", true
			}
			return "Service is running...", true
		}
	}
	return "Service is stopped", false
}

func (linux *Daemon) Install(args ...string) (string, error) {
	installAction := "Install " + linux.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return installAction + failed, err
	}
	if linux.isInstalled() {
		return installAction + failed, ErrAlreadyInstalled
	}
	srvPath := linux.servicePath()
	if linux.template != "" {
		linux.template = systemDConfig
	}
	if err := templateParse("systemVConfig", linux.GetTemplate(), srvPath, templateData{
		Name:         linux.name,
		Description:  linux.desc,
		Author:       linux.author,
		Dependencies: strings.Join(linux.deps, " "),
		Args:         strings.Join(args, " "),
	}); err != nil {
		return installAction + failed, err
	}
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return installAction + failed, err
	}
	if err := exec.Command("systemctl", "enable", linux.name+".service").Run(); err != nil {
		return installAction + failed, err
	}
	return installAction + success, nil
}

// Remove the service
func (linux *Daemon) Remove() (string, error) {
	removeAction := "Removing " + linux.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return removeAction + failed, err
	}
	if !linux.isInstalled() {
		return removeAction + failed, ErrNotInstalled
	}
	if err := exec.Command("systemctl", "disable", linux.name+".service").Run(); err != nil {
		return removeAction + failed, err
	}
	if err := os.Remove(linux.servicePath()); err != nil {
		return removeAction + failed, err
	}
	return removeAction + success, nil
}

// Start the service
func (linux *Daemon) Start() (string, error) {
	startAction := "Starting " + linux.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return startAction + failed, err
	}
	if !linux.isInstalled() {
		return startAction + failed, ErrNotInstalled
	}
	if _, ok := linux.checkRunning(); ok {
		return startAction + failed, ErrAlreadyRunning
	}
	if err := exec.Command("systemctl", "start", linux.name+".service").Run(); err != nil {
		return startAction + failed, err
	}
	return startAction + success, nil
}

// Stop the service
func (linux *Daemon) Stop() (string, error) {
	stopAction := "Stopping " + linux.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return stopAction + failed, err
	}
	if !linux.isInstalled() {
		return stopAction + failed, ErrNotInstalled
	}
	if _, ok := linux.checkRunning(); ok {
		return stopAction + failed, ErrAlreadyRunning
	}
	if err := exec.Command("systemctl", "stop", linux.name+".service").Run(); err != nil {
		return stopAction + failed, err
	}
	return stopAction + success, nil
}

// Start the service
func (linux *Daemon) Restart() (string, error) {
	restartAction := "Restarting " + linux.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return restartAction + failed, err
	}
	if !linux.isInstalled() {
		return restartAction + failed, ErrNotInstalled
	}
	if _, ok := linux.checkRunning(); ok {
		return restartAction + failed, ErrAlreadyRunning
	}
	if err := exec.Command("systemctl", "restart", linux.name+".service").Run(); err != nil {
		return restartAction + failed, err
	}
	return restartAction + success, nil
}

// Status - Get service status
func (linux *Daemon) Status() (string, error) {
	if ok, err := checkPrivileges(); !ok {
		return "", err
	}
	if !linux.isInstalled() {
		return statNotInstalled, ErrNotInstalled
	}
	statusAction, _ := linux.checkRunning()
	return statusAction, nil
}

func (linux *Daemon) GetTemplate() string {
	if linux.template == "" {
		linux.template = systemDConfig
	}
	return linux.template
}

func (linux *Daemon) SetTemplate(tpl string) error {
	linux.template = tpl
	return nil
}
