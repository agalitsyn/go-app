package api

import (
	"net"
	"net/http"

	"github.com/agalitsyn/goapi/handlers"
	"github.com/braintree/manners"
	"github.com/julienschmidt/httprouter"
)

// Provides server and router in conjunction
type API struct {
	Server *manners.GracefulServer
	Router http.Handler
}

// Constructor for API struct
func New(host, port string) *API {
	httpAddr := net.JoinHostPort(host, port)

	router := httprouter.New()
	router.GET("/", handlers.IndexHandler)
	router.GET("/healthz", handlers.HealthzHandler)

	httpServer := manners.NewServer()
	httpServer.Addr = httpAddr
	httpServer.Handler = handlers.LoggingHandler(router)

	return &API{
		Server: httpServer,
		Router: router,
	}
}
