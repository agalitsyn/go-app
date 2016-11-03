// Package main provides web API
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/gravitational/trace"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	app := NewWebApp()

	kingpin.Flag("log-level", "Log level.").Default("info").Envar(EnvLogLevel).StringVar(&app.config.LogLevel)
	kingpin.Flag("host", "HTTP host.").Default("127.0.0.1").Envar(EnvHost).StringVar(&app.config.Host)
	kingpin.Flag("port", "HTTP port.").Default("5000").Envar(EnvPort).StringVar(&app.config.Port)
	kingpin.Flag("tls-cert", "Path to the client server TLS cert file.").Envar(EnvTLSCert).StringVar(&app.config.TLSCert)
	kingpin.Flag("tls-key", "Path to the client server TLS key file.").Envar(EnvTLSKey).StringVar(&app.config.TLSKey)
	kingpin.Flag("tls-ca-cert", "Path to the CAs cert file.").Envar(EnvTLSCACert).StringVar(&app.config.TLSCACert)
	kingpin.Parse()

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
