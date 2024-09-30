package daemon

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/coreos/go-systemd/v22/unit"
)

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
	"Install": {
		"DefaultInstance": "default",
		"WantedBy":        "multi-user.target",
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

func CreateUnit(multi bool, binName, desc, path string, args ...string) ([]byte, error) {
	if multi {
		binName += "@%i"
	}
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
		if multi {
			unitConfig["Service"]["ExecStart"] = path + " --instance %i " + strings.Join(args, " ")
		}else{
			unitConfig["Service"]["ExecStart"] = path + " " + strings.Join(args, " ")
		}
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
