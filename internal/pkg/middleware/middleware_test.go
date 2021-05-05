package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

func TestApiVersion(t *testing.T) {
	r := chi.NewRouter()
	r.Use(APIVersion("v1"))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		v, ok := r.Context().Value(APIVersionContextKey).(string)
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
