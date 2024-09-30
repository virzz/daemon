package daemon_test

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
	"github.com/virzz/vlog"
)

const (
	appID       = "com.virzz.myservice"
	name        = "myservice"
	description = "MyTestService"
)

type Config struct {
	A string `json:"a"`
}

var (
	Version string = "1.0.0"
	Commit  string = "dev"

	C = &Config{}
)

func Action(cmd *cobra.Command, args []string) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-time.After(2 * time.Second):
			log.Println("Myservice is running...")
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

func Example() {
	_, err := daemon.New(appID, name, description, Version, Commit)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	vlog.New("test.log")
	daemon.SetLogger(vlog.Log)

	daemon.EnableRemoteConfig("test")

	daemon.SetUnitConfig("Service", "Type", "simple")
	daemon.RegisterConfig(C)

	err = daemon.ExecuteE(Action)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
