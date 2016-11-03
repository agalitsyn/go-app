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

	log "github.com/Sirupsen/logrus"
)

type WebApp struct {
	config  Config
	router  *httprouter.Router
	server  *manners.GracefulServer
	errChan chan error
}

func NewWebApp() *WebApp {
	router := httprouter.New()
	router.GET("/", IndexHandler)
	router.GET("/healthz", HealthzHandler)

	return &WebApp{
		router:  router,
		server:  manners.NewServer(),
		errChan: make(chan error, 1),
	}
}

func (app *WebApp) Start() error {
	if err := app.config.SetupLogging(); err != nil {
		return trace.Wrap(err)
	}

	log.Infof("Start with config: %+v", app.config)

	app.server.Addr = net.JoinHostPort(app.config.Host, app.config.Port)
	app.server.Handler = handlers.LoggingHandler(os.Stdout, app.router)

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
