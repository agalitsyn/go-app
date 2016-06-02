package handlers

import (
	"net/http"

	"github.com/agalitsyn/goapi/log"
)

func LoggingHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := map[string]interface{}{
			"remote_ip":       r.RemoteAddr,
			"http_uri":        r.URL.Path,
			"http_referer":    r.Referer(),
			"http_user_agent": r.UserAgent(),
			"http_method":     r.Method,
			"http_proto":      r.Proto,
		}
		accessLogger := log.GetLoggerWithFields("api", fields)
		accessLogger.Info("Incoming request")

		h.ServeHTTP(w, r)
	})
}
