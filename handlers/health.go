package handlers

import (
	"net/http"

	"github.com/agalitsyn/goapi/health"
	"github.com/julienschmidt/httprouter"
)

// HealthzHandler responds to health check requests.
func HealthzHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(health.HealthzStatus())
}
