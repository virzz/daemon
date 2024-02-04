package daemon

import (
	"os"
	"os/exec"
	"os/user"
)

const (
	success          = "\t\t\t[  \033[32mOK\033[0m  ]" // Show colored "OK"
	failed           = "\t\t\t[\033[31mFAILED\033[0m]" // Show colored "FAILED"
	statNotInstalled = "Service not installed"
)

// Lookup path for executable file
func executablePath(name ...string) (string, error) {
	if len(name) > 0 {
		if path, err := exec.LookPath(name[0]); err == nil {
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}
	return os.Executable()
}

// Check root rights to use system service
func checkPrivileges() (bool, error) {
	_user, err := user.Current()
	if err != nil {
		return false, ErrUnsupportedSystem
	}
	if _user.Gid == "0" {
		return true, nil
	} else {
		return false, ErrRootPrivileges
	}
}
