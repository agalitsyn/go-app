package service

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"syscall"

	"github.com/agalitsyn/goapi/api"
	"github.com/agalitsyn/goapi/db"
	"github.com/agalitsyn/goapi/health"
	"github.com/agalitsyn/goapi/log"
	"github.com/agalitsyn/goapi/preferences"
	"github.com/pkg/errors"
)

// Kind of singletone godlike object for service with all dependencies
type Service struct {
	log         log.Logger
	api         *api.API
	preferences *preferences.Preferences
	database    *db.Database
	errChan     chan error
	signalChan  chan os.Signal
}

func Start() error {
	p, err := preferences.Get()
	if err != nil {
		return errors.Wrap(err, "Failed create preferences")
	}

	log := log.GetLogger(p.LogFormat, p.LogLevel)
	log.Infof("Got preferences '%#v'", p)

	db, err := db.New(p.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "Can't create database")
	}
	err = db.Connect()
	if err != nil {
		return errors.Wrap(err, "Can't connect to database")
	}
	log.Infof("Connection to database '%v' established", p.DatabaseURL)

	api := api.New("", p.Port)
	log.Infof("HTTP server configured for listening on %v", p.Port)

	service := &Service{
		log:         log,
		api:         api,
		database:    db,
		preferences: p,
		errChan:     make(chan error, 1),
		signalChan:  make(chan os.Signal, 1),
	}

	log.Info("Starting server...")
	service.start()
	return nil
}

func (s *Service) start() {
	go func() {
		s.errChan <- s.api.Server.ListenAndServe()
	}()

	signal.Notify(s.signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-s.errChan:
			if err != nil {
				s.log.WithError(err).Error("Got error from server")
				os.Exit(1)
			}
		case sig := <-s.signalChan:
			s.log.Infof(fmt.Sprintf("Captured %v. Gracefull shutdown...", sig))
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
	health.SetHealthzStatus(http.StatusServiceUnavailable)
	s.database.Close()
	s.api.Server.BlockingClose()
}
