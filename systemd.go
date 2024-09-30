package daemon

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	systemd "github.com/coreos/go-systemd/v22/dbus"
)

type Systemd struct {
	logger      *slog.Logger
	Name        string
	Description string
	Version     string
	AppID       string
}

func (s *Systemd) Install(multi bool, args ...string) error {
	s.logger.Info("Install... " + s.Name)
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	var buf []byte
	buf, err = CreateUnit(multi, s.Name, s.Description, execPath, args...)
	if err != nil {
		return err
	}
	os.WriteFile("/etc/systemd/system/"+s.Name+"@.service", buf, 0644)
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	s.logger.Info("Installed " + s.Name)
	return conn.ReloadContext(ctx)
}

// Remove the service
func (s *Systemd) Remove() error {
	s.logger.Info("Removing... " + s.Name)
	err := s.Stop(true)
	if err != nil {
		s.logger.Warn(err.Error())
	}
	err = os.Remove("/etc/systemd/system/" + s.Name + "@.service")
	if err != nil {
		return err
	}
	s.logger.Info("Removed " + s.Name)
	return nil
}

// Start the service
func (s *Systemd) Start(num int, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	recv := make(chan string, 1)
	if num > 0 {
		for i := 1; i <= num; i++ {
			name := s.Name + "@" + strconv.Itoa(i) + ".service"
			_, err = conn.StartUnitContext(ctx, name, "fail", recv)
			if err != nil {
				s.logger.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				s.logger.Error("Started [ " + name + " ] " + v)
			} else {
				s.logger.Info("Started [ " + name + " ] " + v)
			}
		}
	} else if len(tags) > 0 {
		for _, tag := range tags {
			name := s.Name + "@" + tag + ".service"
			_, err = conn.StartUnitContext(ctx, name, "fail", recv)
			if err != nil {
				s.logger.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				s.logger.Error("Started [ " + name + " ] " + v)
			} else {
				s.logger.Info("Started [ " + name + " ] " + v)
			}
		}
	} else {
		name := s.Name + "@default.service"
		_, err = conn.StartUnitContext(ctx, name, "fail", recv)
		if err != nil {
			return err
		}
		v := <-recv
		if v == "failed" {
			s.logger.Error("Started [ " + name + " ] " + v)
		} else {
			s.logger.Info("Started [ " + name + " ] " + v)
		}
	}
	return nil
}

// Stop the service
func (s *Systemd) Stop(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := s.Status(false)
		if err != nil {
			return err
		}
		recv := make(chan string, 1)
		for _, item := range items {
			_, err = conn.StopUnitContext(ctx, item.Name, "fail", recv)
			if err != nil {
				s.logger.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				s.logger.Error("Stop [ " + item.Name + "] " + v)
			} else {
				s.logger.Info("Stop [ " + item.Name + " ] " + v)
			}
		}
	} else if len(tags) > 0 {
		recv := make(chan string, 1)
		for _, tag := range tags {
			name := s.Name + "@" + tag + ".service"
			_, err = conn.StopUnitContext(ctx, name, "fail", recv)
			if err != nil {
				s.logger.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				s.logger.Error("Stop [" + name + "] " + v)
			} else {
				s.logger.Info("Stop [ " + name + " ] " + v)
			}
		}
	} else {
		recv := make(chan string, 1)
		name := s.Name + "@default.service"
		_, err = conn.StopUnitContext(ctx, name, "fail", recv)
		if err != nil {
			return err
		}
		v := <-recv
		if v == "failed" {
			s.logger.Error("Stop [" + name + "] " + v)
		} else {
			s.logger.Info("Stop [ " + name + " ] " + v)
		}
	}
	return nil
}

// Kill the service
func (s *Systemd) Kill(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := s.Status(false)
		if err != nil {
			return err
		}
		for _, item := range items {
			conn.KillUnitContext(ctx, item.Name, 9)
		}
	} else if len(tags) > 0 {
		for _, tag := range tags {
			conn.KillUnitContext(ctx, s.Name+"@"+tag+".service", 9)
		}
	} else {
		conn.KillUnitContext(ctx, s.Name+"@default.service", 9)
	}
	return nil
}

// Restart the service
func (s *Systemd) Restart(all bool, tags ...string) error {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := s.Status(false)
		if err != nil {
			return err
		}
		recv := make(chan string, 1)
		for _, item := range items {
			_, err = conn.RestartUnitContext(ctx, item.Name, "fail", recv)
			if err != nil {
				s.logger.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				s.logger.Error("Restarted [ " + item.Name + "] " + v)
			} else {
				s.logger.Info("Restarted [ " + item.Name + " ] " + v)
			}
		}
	} else if len(tags) > 0 {
		recv := make(chan string, 1)
		for _, tag := range tags {
			name := s.Name + "@" + tag + ".service"
			_, err = conn.RestartUnitContext(ctx, name, "fail", recv)
			if err != nil {
				s.logger.Warn(err.Error())
				continue
			}
			v := <-recv
			if v == "failed" {
				s.logger.Error("Restarted [ " + name + " ] " + v)
			} else {
				s.logger.Info("Restarted [ " + name + " ] " + v)
			}
		}
	} else {
		recv := make(chan string, 1)
		name := s.Name + "@default.service"
		_, err = conn.RestartUnitContext(ctx, name, "fail", recv)
		if err != nil {
			return err
		}
		v := <-recv
		if v == "failed" {
			s.logger.Error("Restarted [ " + name + " ] " + v)
		} else {
			s.logger.Info("Restarted [ " + name + " ] " + v)
		}
	}
	return nil
}

// Reload the service
func (s *Systemd) Reload(all bool, tags ...string) error {
	s.logger.Info("Reloading... " + s.Name)
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	if all {
		items, err := s.Status(false)
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
				s.logger.Error("Reloaded [ " + item.Name + "] " + v)
			} else {
				s.logger.Info("Reloaded [ " + item.Name + " ] " + v)
			}
		}
	} else if len(tags) > 0 {
		recv := make(chan string, 1)
		for _, tag := range tags {
			name := s.Name + "@" + tag + ".service"
			_, err = conn.ReloadOrRestartUnitContext(ctx, name, "fail", recv)
			if err != nil {
				s.logger.Warn(err.Error())
			}
			v := <-recv
			if v == "failed" {
				s.logger.Error("Reloaded [ " + name + " ] " + v)
			} else {
				s.logger.Info("Reloaded [ " + name + " ] " + v)
			}
		}
	} else {
		recv := make(chan string, 1)
		name := s.Name + "@default.service"
		_, err = conn.ReloadOrRestartUnitContext(ctx, name, "fail", recv)
		if err != nil {
			return err
		}
		v := <-recv
		if v == "failed" {
			s.logger.Error("Reloaded [ " + name + " ] " + v)
		} else {
			s.logger.Info("Reloaded [ " + name + " ] " + v)
		}
	}
	return nil
}

// Status - Get service status
func (s *Systemd) Status(show bool) ([]systemd.UnitStatus, error) {
	ctx := context.Background()
	conn, err := systemd.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := conn.ListUnitsByPatternsContext(ctx, nil, []string{s.Name + "*"})
	if err != nil {
		return nil, err
	}
	if show {
		for _, item := range items {
			if item.SubState == "running" {
				s.logger.Info(item.Name, item.ActiveState, item.SubState)
			} else {
				s.logger.Warn(item.Name, item.ActiveState, item.SubState)
			}
		}
	}
	return items, nil
}
