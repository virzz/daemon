package daemon

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var std *Daemon

// New - Create a new daemon
func New(name, desc, version, author string, deps ...string) (*Daemon, error) {
	std = &Daemon{name: name, desc: desc, version: version, author: author, deps: deps}
	return wrapCmd(std), nil
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
func (r *Daemon) servicePath() string {
	return "/etc/systemd/system/" + r.name + ".service"
}

func (r *Daemon) isInstalled() bool {
	_, err := os.Stat(r.servicePath())
	return err == nil
}

func (r *Daemon) checkRunning() (string, bool) {
	output, err := exec.Command("systemctl", "status", r.name+".service").Output()
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

func (r *Daemon) Install(args ...string) (string, error) {
	installAction := "Install " + r.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return installAction + failed, err
	}
	if r.isInstalled() {
		return installAction + failed, ErrAlreadyInstalled
	}
	srvPath := r.servicePath()
	if r.template != "" {
		r.template = systemDConfig
	}
	if err := templateParse("systemVConfig", r.GetTemplate(), srvPath, templateData{
		Name:         r.name,
		Description:  r.desc,
		Author:       r.author,
		Dependencies: strings.Join(r.deps, " "),
		Args:         strings.Join(args, " "),
	}); err != nil {
		return installAction + failed, err
	}
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return installAction + failed, err
	}
	if err := exec.Command("systemctl", "enable", r.name+".service").Run(); err != nil {
		return installAction + failed, err
	}
	return installAction + success, nil
}

// Remove the service
func (r *Daemon) Remove() (string, error) {
	removeAction := "Removing " + r.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return removeAction + failed, err
	}
	if !r.isInstalled() {
		return removeAction + failed, ErrNotInstalled
	}
	if err := exec.Command("systemctl", "disable", r.name+".service").Run(); err != nil {
		return removeAction + failed, err
	}
	if err := os.Remove(r.servicePath()); err != nil {
		return removeAction + failed, err
	}
	return removeAction + success, nil
}

// Start the service
func (r *Daemon) Start() (string, error) {
	startAction := "Starting " + r.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return startAction + failed, err
	}
	if !r.isInstalled() {
		return startAction + failed, ErrNotInstalled
	}
	if _, ok := r.checkRunning(); ok {
		return startAction + failed, ErrAlreadyRunning
	}
	if err := exec.Command("systemctl", "start", r.name+".service").Run(); err != nil {
		return startAction + failed, err
	}
	return startAction + success, nil
}

// Stop the service
func (r *Daemon) Stop() (string, error) {
	stopAction := "Stopping " + r.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return stopAction + failed, err
	}
	if !r.isInstalled() {
		return stopAction + failed, ErrNotInstalled
	}
	if _, ok := r.checkRunning(); !ok {
		return stopAction + failed, ErrAlreadyStopped
	}
	if err := exec.Command("systemctl", "stop", r.name+".service").Run(); err != nil {
		return stopAction + failed, err
	}
	return stopAction + success, nil
}

// Start the service
func (r *Daemon) Restart() (string, error) {
	restartAction := "Restarting " + r.desc + ":"
	if ok, err := checkPrivileges(); !ok {
		return restartAction + failed, err
	}
	if !r.isInstalled() {
		return restartAction + failed, ErrNotInstalled
	}
	if _, ok := r.checkRunning(); !ok {
		return restartAction + failed, ErrAlreadyStopped
	}
	if err := exec.Command("systemctl", "restart", r.name+".service").Run(); err != nil {
		return restartAction + failed, err
	}
	return restartAction + success, nil
}

// Status - Get service status
func (r *Daemon) Status() (string, error) {
	if ok, err := checkPrivileges(); !ok {
		return "", err
	}
	if !r.isInstalled() {
		return statNotInstalled, ErrNotInstalled
	}
	statusAction, _ := r.checkRunning()
	return statusAction, nil
}

func (r *Daemon) GetTemplate() string {
	if r.template == "" {
		r.template = systemDConfig
	}
	return r.template
}

func (r *Daemon) SetTemplate(tpl string) error {
	r.template = tpl
	return nil
}
