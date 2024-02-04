package daemon

import (
	_ "embed"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed systemd.service
var systemDConfig string

type templateData struct {
	Name, Description, Author, Dependencies, WorkDir, Path, Args string
}

func templateParse(name, content, srvPath string, td templateData) error {
	execPath, err := executablePath(td.Name)
	if err != nil {
		return err
	}
	td.Path = execPath
	td.WorkDir = filepath.Dir(execPath)
	templ, err := template.New(name).Parse(content)
	if err != nil {
		return err
	}
	file, err := os.Create(srvPath)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := templ.Execute(file, &td); err != nil {
		return err
	}
	return nil
}
