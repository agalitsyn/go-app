package handler

import (
	"net/http"

	"github.com/agalitsyn/goapi/pkg/log"

	"github.com/goware/cors"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/pressly/chi/render"
)

// Router represents HTTP route multiplexer
type Router struct {
	*chi.Mux
}

// New creates a new http.Hander with the provided options
func New(opts ...Option) *Router {
	r := &Router{chi.NewRouter()}
	r.Use(render.SetContentType(render.ContentTypeJSON))

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Option represents Router option
type Option func(*Router)

// WithLogging adds requests logging middleware (aka access log)
func WithLogging(logger log.Logger) Option {
	return func(r *Router) { r.Use(newStructuredLogger(logger)) }
}

// WithRequestID adds middleware which adds unique request id
func WithRequestID() Option {
	return func(r *Router) { r.Use(middleware.RequestID) }
}

// WithRealIP adds middleware which helps get real requester's IP, not proxy
func WithRealIP() Option {
	return func(r *Router) { r.Use(middleware.RealIP) }
}

// WithRecover adds recover middleware, which can catch panics from handlers
func WithRecover() Option {
	return func(r *Router) { r.Use(middleware.Recoverer) }
}

// WithCORS adds cross-domain request setup
func WithCORS(allowedOrigins, allowedHeaders, exposedHeaders []string) Option {
	return func(r *Router) {
		m := cors.New(cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowedHeaders:   allowedHeaders,
			ExposedHeaders:   exposedHeaders,
			AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodDelete},
			AllowCredentials: true,
		})
		r.Use(m.Handler)
	}
}

func newStructuredLogger(logger log.Logger) func(next http.Handler) http.Handler {
	l, ok := logger.(middleware.LogFormatter)
	if !ok {
		return middleware.Logger
	}
	return middleware.RequestLogger(l)
}
