// Package web provides web application
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/handlers"
	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/braintree/manners"
)

func main() {
	var cfg Config

	kingpin.Flag("log-level", "Log level.").Default("info").Envar(EnvLogLevel).StringVar(&cfg.LogLevel)
	kingpin.Flag("host", "HTTP host.").Default("127.0.0.1").Envar(EnvHost).StringVar(&cfg.Host)
	kingpin.Flag("port", "HTTP port.").Default("5000").Envar(EnvPort).StringVar(&cfg.Port)
	kingpin.Flag("database-url", "Database URL for connection.").Envar(EnvDatabaseURL).StringVar(&cfg.DatabaseURL)
	kingpin.Flag("tls-cert", "Path to the client server TLS cert file.").Envar(EnvTLSCert).StringVar(&cfg.TLSCert)
	kingpin.Flag("tls-key", "Path to the client server TLS key file.").Envar(EnvTLSKey).StringVar(&cfg.TLSKey)
	kingpin.Flag("tls-ca-cert", "Path to the CAs cert file.").Envar(EnvTLSCACert).StringVar(&cfg.TLSCACert)
	kingpin.Parse()

	if err := cfg.SetupLogging(); err != nil {
		log.Fatal(trace.DebugReport(err))
	}

	log.Infof("Start with config: %+v", cfg)

	db, err := GetDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(trace.DebugReport(err))
	}

	log.Debugf("Connecting to database at %s", cfg.DatabaseURL)
	if err = db.Connect(); err != nil {
		log.Fatal(trace.DebugReport(err))
	}

	caCert, err := ioutil.ReadFile(cfg.TLSCACert)
	if err != nil {
		log.Fatal(trace.DebugReport(err))
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()

	router := httprouter.New()
	router.GET("/", IndexHandler)
	router.GET("/healthz", HealthzHandler)

	httpServer := manners.NewServer()
	httpServer.Addr = net.JoinHostPort(cfg.Host, cfg.Port)
	httpServer.Handler = handlers.LoggingHandler(os.Stdout, router)
	httpServer.TLSConfig = tlsConfig

	errChan := make(chan error, 10)
	go func() {
		errChan <- httpServer.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey)
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Fatal(trace.DebugReport(err))
			}
		case s := <-signalChan:
			log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
			httpServer.BlockingClose()
			os.Exit(0)
		}
	}
}
