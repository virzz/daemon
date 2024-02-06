package main

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/manifoldco/promptui"
)

var (
	//go:embed daemon.tmpl
	daemonTmpl string

	//go:embed action.tmpl
	actionTmpl string
)

type Opt struct {
	Name, Description, Author string
	Signal                    bool
}

func main() {
	opt := &Opt{}
	prompt := promptui.Prompt{Label: "Name"}
	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	opt.Name = result

	prompt = promptui.Prompt{Label: "Description"}
	result, err = prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	opt.Description = result

	prompt = promptui.Prompt{
		Label:   "Author",
		Default: "陌竹 <mozhu233@outlook.com>",
	}
	result, err = prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	opt.Author = result

	prompt = promptui.Prompt{
		Label:   "Signal",
		Default: "",
	}
	result, err = prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	if result != "" {
		opt.Signal = true
	}

	f, err := os.Create("main.go")
	if err != nil {
		fmt.Printf("Create file failed %v\n", err)
		return
	}
	err = template.Must(template.New("daemon").Parse(daemonTmpl)).Execute(f, opt)
	if err != nil {
		fmt.Printf("Execute template failed %v\n", err)
		return
	}
	f.Close()
	f, err = os.Create("action.go")
	if err != nil {
		fmt.Printf("Create file failed %v\n", err)
		return
	}
	defer f.Close()
	err = template.Must(template.New("daemon").Parse(actionTmpl)).Execute(f, opt)
	if err != nil {
		fmt.Printf("Execute template failed %v\n", err)
		return
	}
}
