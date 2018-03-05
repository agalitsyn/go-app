package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agalitsyn/goapi/pkg/log"
)

func TestApiVersion(t *testing.T) {
	r := New(WithLogging(log.New("", "", ioutil.Discard)))
	r.Use(ApiVersion("v1"))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		v, ok := r.Context().Value(ApiVersionContextKey).(string)
		if !ok {
			t.Fatal("cast failed")
		}
		w.Write([]byte(v))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	r.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "v1" {
		t.Fatalf("expected v1, got %s", string(body))
	}
}
