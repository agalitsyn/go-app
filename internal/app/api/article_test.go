package api

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agalitsyn/go-app/internal/storage"

	"github.com/go-chi/chi"
)

func TestArticleService_listHandler(t *testing.T) {
	t.Parallel()

	store := newMockArticleStorage(map[int]storage.Article{
		1: {ID: 1, Title: "Foo", Slug: "foo"},
	})
	service := NewArticleService(store)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	r := chi.NewRouter()
	r.Get("/", service.listHandler)
	r.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotEmpty(t, body)
}

func TestArticleService_deleteHandler(t *testing.T) {
	t.Parallel()

	service := NewArticleService(&mockArticleStorage{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/1", nil)

	r := chi.NewRouter()
	r.Delete("/{id}", service.deleteHandler)
	r.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestArticleService_storeHandler(t *testing.T) {
	t.Parallel()

	service := NewArticleService(&mockArticleStorage{})

	r := chi.NewRouter()
	r.Post("/{id}", service.storeHandler)

	tests := []struct {
		name    string
		payload string
		code    int
	}{
		{
			name:    "invalid payload",
			payload: `foo`,
			code:    http.StatusBadRequest,
		},
		{
			name:    "empty",
			payload: `{}`,
			code:    http.StatusBadRequest,
		},
		{
			name: "no slug",
			payload: `{
				"title": "New",
			}`,
			code: http.StatusBadRequest,
		},
		{
			name: "no title",
			payload: `{
				"slug": "new",
			}`,
			code: http.StatusBadRequest,
		},
		{
			name: "create",
			payload: `{
				"title": "New",
				"slug": "new"
			}`,
			code: http.StatusOK,
		},
		{
			name: "update",
			payload: `{
				"title": "Not new",
				"slug": "not-new"
			}`,
			code: http.StatusOK,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBufferString(tt.payload)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "http://example.com/1", buf)

			r.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, tt.code, resp.StatusCode)
		})
	}
}

type mockArticleStorage struct {
	data map[int]storage.Article
}

func newMockArticleStorage(data map[int]storage.Article) *mockArticleStorage {
	return &mockArticleStorage{
		data: data,
	}
}

func (s *mockArticleStorage) FilterArticles(ctx context.Context, filter storage.ArticleFilter) ([]storage.Article, error) {
	var res []storage.Article
	for _, article := range s.data {
		res = append(res, article)
	}
	return res, nil
}

func (s *mockArticleStorage) StoreArticles(ctx context.Context, articles []storage.Article) error {
	return nil
}

func (s *mockArticleStorage) DeleteArticles(ctx context.Context, articles []storage.Article) error {
	return nil
}
