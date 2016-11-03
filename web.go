package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/braintree/manners"
	"github.com/gorilla/handlers"
	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"

	"gopkg.in/alecthomas/kingpin.v2"

	log "github.com/Sirupsen/logrus"
)

type WebApp struct {
	cli     *kingpin.Application
	server  *manners.GracefulServer
	config  Config
	errChan chan error
}

func NewWebApp() *WebApp {
	return &WebApp{
		cli:     kingpin.New("goapi", ""),
		server:  manners.NewServer(),
		errChan: make(chan error, 1),
	}
}

func (app *WebApp) Start() error {
	app.cli.Flag("log-level", "Log level.").Default("info").Envar(EnvLogLevel).StringVar(&app.config.LogLevel)
	app.cli.Flag("host", "HTTP host.").Default("127.0.0.1").Envar(EnvHost).StringVar(&app.config.Host)
	app.cli.Flag("port", "HTTP port.").Default("5000").Envar(EnvPort).StringVar(&app.config.Port)
	app.cli.Flag("tls-cert", "Path to the client server TLS cert file.").Envar(EnvTLSCert).StringVar(&app.config.TLSCert)
	app.cli.Flag("tls-key", "Path to the client server TLS key file.").Envar(EnvTLSKey).StringVar(&app.config.TLSKey)
	app.cli.Flag("tls-ca-cert", "Path to the CAs cert file.").Envar(EnvTLSCACert).StringVar(&app.config.TLSCACert)
	if _, err := app.cli.Parse(os.Args[1:]); err != nil {
		return trace.Wrap(err)
	}

	if err := app.config.SetupLogging(); err != nil {
		return trace.Wrap(err)
	}

	log.Infof("Start with config: %+v", app.config)

	router := httprouter.New()
	router.GET("/", IndexHandler)
	router.GET("/healthz", HealthzHandler)

	app.server.Addr = net.JoinHostPort(app.config.Host, app.config.Port)
	app.server.Handler = handlers.LoggingHandler(os.Stdout, router)

	// HTTPS
	if app.config.TLSCert != "" && app.config.TLSKey != "" {
		if app.config.TLSCACert != "" {
			caCert, err := ioutil.ReadFile(app.config.TLSCACert)
			if err != nil {
				return trace.Wrap(err, "Can't read CA file: %v", app.config.TLSCACert)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig := &tls.Config{
				ClientCAs:  caCertPool,
				ClientAuth: tls.RequireAndVerifyClientCert,
			}
			tlsConfig.BuildNameToCertificate()

			app.server.TLSConfig = tlsConfig
		}

		go func() {
			app.errChan <- app.server.ListenAndServeTLS(app.config.TLSCert, app.config.TLSKey)
		}()
	} else {
		// HTTP
		go func() {
			app.errChan <- app.server.ListenAndServe()
		}()
	}

	if err := <-app.errChan; err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func (app *WebApp) Stop() {
	log.Info("Stop")
	SetHealthzStatus(http.StatusServiceUnavailable)
	app.server.BlockingClose()
}
