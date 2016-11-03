package main

import (
	"os"
	"strings"

	"github.com/gravitational/trace"

	log "github.com/Sirupsen/logrus"
)

const (
	EnvLogLevel  = "LOG_LEVEL"
	EnvHost      = "HOST"
	EnvPort      = "PORT"
	EnvTLSCert   = "TLS_CERT"
	EnvTLSKey    = "TLS_KEY"
	EnvTLSCACert = "TLS_CA_CERT"
)

type Config struct {
	LogLevel  string
	Host      string
	Port      string
	TLSCert   string
	TLSKey    string
	TLSCACert string
}

func (c *Config) SetupLogging() error {
	// clear existing hooks:
	log.StandardLogger().Hooks = make(log.LevelHooks)

	level := strings.ToLower(c.LogLevel)
	if level == "debug" {
		trace.SetDebug(true)
	}
	sev, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(sev)

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	return nil
}
