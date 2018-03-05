package health

import (
	"net/http"

	"github.com/go-chi/chi"
)

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", readinessHandler)
	return r
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(ReadinessStatus())
}
