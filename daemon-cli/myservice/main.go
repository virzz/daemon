package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/virzz/daemon/v2"
)

var (
	name        = "myservice"
	version     = "1.0.0"
	description = "My Test Service"
)

func Action(cmd *cobra.Command, args []string) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-time.After(1 * time.Second):
			log.Println("Running...")
		case killSignal := <-interrupt:
			fmt.Println("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				fmt.Println("Daemon was killed")
				return nil
			}
			return nil
		}
	}
}

func main() {
	_, err := daemon.New(name, description, version)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	daemon.Execute(Action)
}
