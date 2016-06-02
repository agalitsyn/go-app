package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type JSONResponse struct {
	Message string `json:"message"`
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Unfortunately we should show httprouter dependency at func signature
	response := JSONResponse{
		Message: "Hello world",
	}
	json.NewEncoder(w).Encode(response)
	return
}
