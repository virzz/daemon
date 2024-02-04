package daemon

import "github.com/pkg/errors"

var (
	ErrUnsupportedSystem = errors.New("Unsupported system")
	ErrRootPrivileges    = errors.New("You must have root user privileges. Possibly using 'sudo' command should help")
	ErrAlreadyInstalled  = errors.New("Service has already been installed")
	ErrNotInstalled      = errors.New("Service is not installed")
	ErrAlreadyRunning    = errors.New("Service is already running")
	ErrAlreadyStopped    = errors.New("Service has already been stopped")
)
