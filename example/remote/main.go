package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/virzz/daemon/v2"
)

const (
	AppID       = "com.virzz.app.daemon.remote"
	Name        = "srvx"
	Description = "SrvxService"
)

var (
	Version string = "latest"
	Commit  string = "unknown"
)

func Action(cmd *cobra.Command, args []string) error {
	fmt.Println("Hello World")
	return nil
}

func main() {
	daemon.New(AppID, Name, Description, Version, Commit)
	daemon.EnableRemoteConfig("test")
	if err := daemon.ExecuteE(Action); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
