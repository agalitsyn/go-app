package handler

import (
	"context"
	"net/http"
)

type contextKey string

const ApiVersionContextKey contextKey = "api.version"

func ApiVersion(version string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ApiVersionContextKey, version))
			next.ServeHTTP(w, r)
		})
	}
}
