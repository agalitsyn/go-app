package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func doHTTPRequest(method, url string) *http.Response {
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	request, _ := http.NewRequest(method, url, nil)
	response, _ := client.Do(request)
	return response
}

func TestHealthzHandler(t *testing.T) {
	router := httprouter.New()
	router.GET("/healthz", HealthzHandler)

	srv := httptest.NewServer(router)
	defer srv.Close()

	response := doHTTPRequest("GET", fmt.Sprintf("%s/healthz", srv.URL))
	assert.Exactly(t, http.StatusOK, response.StatusCode)
}
