package daemon

import (
	"bytes"
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
	data, err := templateParseData(name, content, td)
	if err != nil {
		return err
	}
	return os.WriteFile(srvPath, data, 0755)
}

func templateParseData(name, content string, td templateData) ([]byte, error) {
	execPath, err := executablePath(td.Name)
	if err != nil {
		return nil, err
	}
	td.Path = execPath
	td.WorkDir = filepath.Dir(execPath)
	templ, err := template.New(name).Parse(content)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if err := templ.Execute(buf, &td); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
