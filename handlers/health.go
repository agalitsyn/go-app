package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/agalitsyn/goapi/health"
)

// HealthzHandler responds to health check requests.
func HealthzHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(health.HealthzStatus())
}
