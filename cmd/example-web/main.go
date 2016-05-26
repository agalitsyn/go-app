package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"strings"

	"github.com/apex/log"

	"github.com/braintree/manners"

	"syscall"

	"database/sql"
	"net"
	"time"

	_ "github.com/lib/pq"

	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
	"github.com/julienschmidt/httprouter"
	"github.com/kelseyhightower/envconfig"
	"github.com/agalitsyn/goapi/handlers"
	"github.com/agalitsyn/goapi/health"
	"github.com/agalitsyn/goapi/preferences"
)

func main() {
	var (
		p  preferences.Preferences
		db *sql.DB
	)

	err := envconfig.Process("goexample", &p)
	if err != nil {
		log.WithError(err).Error("Can't parse config")
	}

	if strings.ToLower(p.LogFormat) == "text" {
		log.SetHandler(text.New(os.Stdout))
	} else {
		log.SetHandler(json.New(os.Stdout))
	}

	lvl, err := log.ParseLevel(p.LogLevel)
	if err != nil {
		log.WithError(err).Error("Log level is invalid")
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)

	// Connect to database.
	log.Infof("Connecting to database at '%v'.", p.DatabaseURL)
	dsn := p.DatabaseURL

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.WithError(err).Error("The data source arguments are not valid")
	}

	var dbError error
	maxAttempts := 30
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		dbError = db.Ping()
		if dbError == nil {
			break
		}
		log.WithError(dbError).Error("Could not establish a connection with the database")
		time.Sleep(time.Duration(attempts) * time.Second)
	}
	if dbError != nil {
		log.WithError(dbError).Fatal("Connection to database failed.")
	}

	httpAddr := net.JoinHostPort("", p.Port)
	log.Info("Starting server...")
	log.Infof("HTTP service listening on %v", httpAddr)

	router := httprouter.New()
	router.GET("/", handlers.IndexHandler)
	router.GET("/healthz", handlers.HealthzHandler)

	httpServer := manners.NewServer()
	httpServer.Addr = httpAddr
	httpServer.Handler = handlers.LoggingHandler(router)

	errChan := make(chan error, 10)
	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.WithError(err).Error("Recieved error")
			}
		case s := <-signalChan:
			log.Infof(fmt.Sprintf("Captured %v. Exiting...", s))
			health.SetHealthzStatus(http.StatusServiceUnavailable)
			httpServer.BlockingClose()
			os.Exit(0)
		}
	}
}
