package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/virzz/daemon"
)

var (
	name         = "myservice"
	version      = "1.0.0"
	description  = "My Test Service"
	author       = "陌竹 <mozhu233@outlook.com>"
	dependencies = []string{""}
)

func Action() error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-time.After(5 * time.Second):
			fmt.Println("Timeout:")
			return nil
		case killSignal := <-interrupt:
			fmt.Println("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return nil
			}
			fmt.Println("Daemon was killed")
			return nil
		}
	}
}

func main() {
	_, err := daemon.New(name, description, version, author, dependencies...)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	daemon.Execute(Action)
}
