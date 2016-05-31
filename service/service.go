package service

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"strings"

	"github.com/apex/log"

	"syscall"

	"database/sql"
	"time"

	_ "github.com/lib/pq"

	"github.com/agalitsyn/goapi/api"
	"github.com/agalitsyn/goapi/health"
	"github.com/agalitsyn/goapi/preferences"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
)

type Service struct {
	preferences *preferences.Preferences
	errChan     chan error
	signalChan  chan os.Signal
}

func (s *Service) start() error {
	// Configure logger
	if strings.ToLower(s.preferences.LogFormat) == "text" {
		log.SetHandler(text.New(os.Stdout))
	} else {
		log.SetHandler(json.New(os.Stdout))
	}
	lvl, err := log.ParseLevel(s.preferences.LogLevel)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)

	// Connect to database.
	log.Infof("Connecting to database at '%v'.", s.preferences.DatabaseURL)
	dsn := s.preferences.DatabaseURL

	var db *sql.DB
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
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
		return dbError
	}

	// Setup HTTP server
	log.Info("Starting server...")
	log.Infof("HTTP service listening on %v", s.preferences.Port)
	api := api.New("", s.preferences.Port)

	// Handle errors and signals
	go func() {
		s.errChan <- api.Server.ListenAndServe()
	}()

	signal.Notify(s.signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-s.errChan:
			if err != nil {
				log.WithError(err).Error("Recieved error")
			}
		case s := <-s.signalChan:
			log.Infof(fmt.Sprintf("Captured %v. Gracefull shutdown...", s))
			health.SetHealthzStatus(http.StatusServiceUnavailable)
			api.Server.BlockingClose()

			switch s {
			case syscall.SIGINT:
				os.Exit(130)
			case syscall.SIGTERM:
				os.Exit(0)
			}
		}
	}
}

func Start() error {
	p, err := preferences.Get()
	if err != nil {
		return err
	}

	service := &Service{
		preferences: p,
		errChan:     make(chan error, 10),
		signalChan:  make(chan os.Signal, 1),
	}

	if err := service.start(); err != nil {
		return err
	}

	return nil
}
