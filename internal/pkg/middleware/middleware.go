package middleware

import (
	"context"
	"net/http"

	"github.com/agalitsyn/go-app/internal/pkg/log"

	"github.com/go-chi/chi/middleware"
)

type contextKey string

const APIVersionContextKey contextKey = "api.version"

func APIVersion(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), APIVersionContextKey, version))
			next.ServeHTTP(w, r)
		})
	}
}

func RequestLogger(logger log.Logger) func(next http.Handler) http.Handler {
	l, ok := logger.(middleware.LogFormatter)
	if !ok {
		return middleware.Logger
	}
	return middleware.RequestLogger(l)
}
