package service

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"strings"

	"github.com/apex/log"

	"syscall"

	"github.com/agalitsyn/goapi/api"
	"github.com/agalitsyn/goapi/db"
	"github.com/agalitsyn/goapi/health"
	"github.com/agalitsyn/goapi/preferences"
	"github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
)

type Service struct {
	api         *api.API
	preferences *preferences.Preferences
	database    *db.Database
	errChan     chan error
	signalChan  chan os.Signal
}

func Start() error {
	p, err := preferences.Get()
	if err != nil {
		return err
	}

	db, err := db.New(p.DatabaseURL)
	if err != nil {
		return err
	}

	err = db.Connect()
	if err != nil {
		return err
	}

	api := api.New("", p.Port)

	service := &Service{
		api:         api,
		database:    db,
		preferences: p,
		errChan:     make(chan error, 10),
		signalChan:  make(chan os.Signal, 1),
	}
	if err := service.start(); err != nil {
		return err
	}
	return nil
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

	// Setup HTTP server
	log.Info("Starting server...")
	log.Infof("HTTP service listening on %v", s.preferences.Port)

	s.serve()
	return nil
}

func (s *Service) serve() {
	// Handle errors and signals
	go func() {
		s.errChan <- s.api.Server.ListenAndServe()
	}()

	signal.Notify(s.signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-s.errChan:
			if err != nil {
				log.WithError(err).Error("Recieved error")
			}
		case sig := <-s.signalChan:
			log.Infof(fmt.Sprintf("Captured %v. Gracefull shutdown...", sig))
			s.stop()

			switch sig {
			case syscall.SIGINT:
				os.Exit(130)
			case syscall.SIGTERM:
				os.Exit(0)
			}
		}
	}
}

func (s *Service) stop() {
	s.database.Close()
	health.SetHealthzStatus(http.StatusServiceUnavailable)
	s.api.Server.BlockingClose()
}
