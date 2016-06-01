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

type Service struct {
	logger      log.Logger
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

	db, err := db.New(p.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "Can't create database")
	}

	err = db.Connect()
	if err != nil {
		return errors.Wrap(err, "Can't connect to database")
	}

	api := api.New("", p.Port)
	log := log.GetLogger(p.LogFormat, p.LogLevel)
	service := &Service{
		logger:      log,
		api:         api,
		database:    db,
		preferences: p,
		errChan:     make(chan error, 10),
		signalChan:  make(chan os.Signal, 1),
	}
	service.start()
	return nil
}

func (s *Service) start() error {
	go func() {
		s.errChan <- s.api.Server.ListenAndServe()
	}()

	signal.Notify(s.signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-s.errChan:
			if err != nil {
				s.logger.WithError(err).Error("Recieved error")
			}
		case sig := <-s.signalChan:
			s.logger.Infof(fmt.Sprintf("Captured %v. Gracefull shutdown...", sig))
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
