package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Hello world!\n"))
}

func HealthzHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(HealthzStatus())
}
