package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/virzz/daemon/v2"
)

const (
	name        = "myservice"
	description = "MyTestService"
)

var (
	version string = "1.0.0"
	commit  string = "dev"
)

func Action(cmd *cobra.Command, args []string) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-time.After(2 * time.Second):
			log.Println("Running...")
			log.Printf("%+v\n", viper.AllSettings())
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
	_, err := daemon.New(name, description, version+" "+commit)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	daemon.SetUnitConfig("Service", "Type", "simple")
	daemon.Execute(Action)
}
