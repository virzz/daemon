package daemon_test

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/virzz/daemon"
)

var (
	name        = "myservice"
	version     = "1.0.0"
	description = "MyService"
	port        = ":9977"
	author      = "陌竹 <mozhu233@outlook.com>"
)

// Action defines daemon actions
func Action() (string, error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return "Possibly was a problem with the port binding", err
	}
	listen := make(chan net.Conn, 100)
	go func(listener net.Listener, listen chan<- net.Conn) {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			listen <- conn
		}
	}(listener, listen)
	for {
		select {
		case conn := <-listen:
			go func(client net.Conn) {
				for {
					buf := make([]byte, 4096)
					numbytes, err := client.Read(buf)
					if numbytes == 0 || err != nil {
						return
					}
					client.Write(buf[:numbytes])
				}
			}(conn)
		case killSignal := <-interrupt:
			listener.Close()
			if killSignal == os.Interrupt {
				return "Daemon was interrupted by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}
}

func Example_myService() {
	_, err := daemon.New(name, description, version, author)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	daemon.Execute(func() error {
		_, err := Action()
		return err
	})
}