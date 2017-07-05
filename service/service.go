package service

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

type Server struct {
	*http.Server
}

func New(addr string, h http.Handler) *Server {
	return &Server{&http.Server{Addr: addr, Handler: h}}
}

func (s *Server) Start() error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.ListenAndServe()
	}()
	if err := <-errChan; err != nil {
		if err != http.ErrServerClosed {
			return errors.Wrap(err, "server error")
		}
	}
	return nil
}

func (s *Server) Stop() error {
	if err := s.Shutdown(context.Background()); err != nil {
		return errors.Wrap(err, "could not shutdown server")
	}
	return nil
}
