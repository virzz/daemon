package main

import (
	_ "embed"
	"log"
	"os"
	"text/template"

	"github.com/manifoldco/promptui"
)

//go:embed daemon.tmpl
var daemonTmpl string

type Opt struct {
	Name, Description string
	Signal            bool
}

func main() {
	opt := &Opt{}
	prompt := promptui.Prompt{Label: "Name"}
	result, err := prompt.Run()
	if err != nil {
		log.Printf("Prompt failed %v\n", err)
		return
	}
	opt.Name = result

	prompt = promptui.Prompt{Label: "Description"}
	result, err = prompt.Run()
	if err != nil {
		log.Printf("Prompt failed %v\n", err)
		return
	}
	opt.Description = result

	prompt = promptui.Prompt{
		Label:   "Signal",
		Default: "",
	}
	result, err = prompt.Run()
	if err != nil {
		log.Printf("Prompt failed %v\n", err)
		return
	}
	if result != "" {
		opt.Signal = true
	}

	f, err := os.Create("main.go")
	if err != nil {
		log.Printf("Create file failed %v\n", err)
		return
	}
	err = template.Must(template.New("daemon").Parse(daemonTmpl)).Execute(f, opt)
	if err != nil {
		log.Printf("Execute template failed %v\n", err)
		return
	}
	f.Close()
}
