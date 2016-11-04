// Package main provides web API
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/gravitational/trace"

	"gopkg.in/alecthomas/kingpin.v2"
)

const EnvLogLevel = "LOG_LEVEL"

func main() {
	app := NewWebApp()

	// Logger is binary-level, not app level. You can start different processes in 1 binary.
	logLevel := kingpin.Flag("log-level", "Log level.").Default("info").Envar(EnvLogLevel).String()
	// Parse this to web app config
	kingpin.Flag("host", "HTTP host.").Default("127.0.0.1").Envar(EnvHost).StringVar(&app.config.Host)
	kingpin.Flag("port", "HTTP port.").Default("5000").Envar(EnvPort).StringVar(&app.config.Port)
	kingpin.Flag("tls-cert", "Path to the client server TLS cert file.").Envar(EnvTLSCert).StringVar(&app.config.TLSCert)
	kingpin.Flag("tls-key", "Path to the client server TLS key file.").Envar(EnvTLSKey).StringVar(&app.config.TLSKey)
	kingpin.Flag("tls-ca-cert", "Path to the CAs cert file.").Envar(EnvTLSCACert).StringVar(&app.config.TLSCACert)
	kingpin.Flag("static-root", "The absolute path to the directory where static files is collected.").Default("/tmp").Envar(EnvStaticRoot).StringVar(&app.config.StaticRoot)
	kingpin.Flag("static-url", "URL to use when referring to static files located in static root.").Default("/static").Envar(EnvStaticURL).StringVar(&app.config.StaticURL)
	kingpin.Parse()

	if err := setupLogging(*logLevel); err != nil {
		log.Fatal(trace.DebugReport(err))
	}

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

func setupLogging(level string) error {
	// clear existing hooks:
	log.StandardLogger().Hooks = make(log.LevelHooks)

	lvl := strings.ToLower(level)
	if lvl == "debug" {
		trace.SetDebug(true)
	}
	sev, err := log.ParseLevel(lvl)
	if err != nil {
		return err
	}
	log.SetLevel(sev)

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	return nil
}
