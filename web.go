package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/braintree/manners"
	"github.com/gorilla/handlers"
	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"

	log "github.com/Sirupsen/logrus"
)

const (
	EnvHost       = "HOST"
	EnvPort       = "PORT"
	EnvTLSCert    = "TLS_CERT"
	EnvTLSKey     = "TLS_KEY"
	EnvTLSCACert  = "TLS_CA_CERT"
	EnvStaticRoot = "STATIC_ROOT"
	EnvStaticURL  = "STATIC_URL"
)

type Config struct {
	Host       string
	Port       string
	TLSCert    string
	TLSKey     string
	TLSCACert  string
	StaticRoot string
	StaticURL  string
}

type WebApp struct {
	config  *Config
	router  *httprouter.Router
	server  *manners.GracefulServer
	errChan chan error
}

func NewWebApp() *WebApp {
	return &WebApp{
		config:  &Config{},
		router:  httprouter.New(),
		server:  manners.NewServer(),
		errChan: make(chan error, 1),
	}
}

func (app *WebApp) Start() error {
	log.Infof("Start with config: %+v", app.config)

	app.router.ServeFiles(fmt.Sprintf("%v/*filepath", app.config.StaticURL), http.Dir(app.config.StaticRoot))
	app.router.PUT(fmt.Sprintf("%v/:name", app.config.StaticURL), StaticMiddleware(app.config.StaticRoot, UpdateHandler))
	app.router.POST(fmt.Sprintf("%v/:name", app.config.StaticURL), StaticMiddleware(app.config.StaticRoot, UploadHandler))
	app.router.DELETE(fmt.Sprintf("%v/:name", app.config.StaticURL), StaticMiddleware(app.config.StaticRoot, DeleteHandler))

	app.router.GET("/healthz", HealthzHandler)

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

func UploadHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sr := r.Context().Value("static_root")
	root, ok := sr.(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// TODO: finish method
	fmt.Printf("root = %+v\n", root)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sr := r.Context().Value("static_root")
	root, ok := sr.(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	// TODO: finist method
	fmt.Printf("root = %+v\n", root)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	sr := r.Context().Value("static_root")
	root, ok := sr.(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	path := filepath.Join(root, ps.ByName("name"))
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		log.Warn(err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err = os.Remove(path); err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func HealthzHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(HealthzStatus())
}

func StaticMiddleware(staticRoot string, h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(r.Context(), "static_root", staticRoot)
		h(w, r.WithContext(ctx), ps)
	}
}
