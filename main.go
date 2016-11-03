// Package main provides web API
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/gravitational/trace"
)

func main() {
	app := NewWebApp()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-signalChan
		log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
		app.Stop()

		switch s {
		case syscall.SIGINT:
			os.Exit(130)
		case syscall.SIGTERM:
			os.Exit(0)
		}
	}()

	if err := app.Start(); err != nil {
		log.Fatal(trace.DebugReport(err))
	}
}
