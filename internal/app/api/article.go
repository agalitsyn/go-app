package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/agalitsyn/go-app/internal/pkg/log"
	"github.com/agalitsyn/go-app/internal/pkg/response"
	"github.com/agalitsyn/go-app/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type ArticleService struct {
	store storage.ArticleRepository
}

func NewArticleService(store storage.ArticleRepository) *ArticleService {
	return &ArticleService{
		store: store,
	}
}

func (s *ArticleService) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", s.listHandler)
	r.Post("/", s.storeHandler)
	r.Route("/{slug}", func(r chi.Router) {
		r.Delete("/", s.deleteHandler)
	})

	return r
}

func (s *ArticleService) listHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.RequestLogger(r)

	articles, err := s.store.FilterArticles(ctx, storage.ArticleFilter{})
	if err != nil {
		logger.WithError(err).Error("could not filter articles")
		response.MustRender(w, r, response.ErrUnknown(err))
		return
	}

	response.MustRenderList(w, r, newArticleListResponse(articles))
}

func (s *ArticleService) storeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.RequestLogger(r)

	var data articleRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.MustRender(w, r, response.ErrBadRequest(err))
		return
	}

	if err := data.Validate(); err != nil {
		response.MustRender(w, r, response.ErrBadRequest(err))
		return
	}

	article := storage.Article{
		Title: data.Title,
		Slug:  data.Slug,
	}

	if err := s.store.StoreArticles(ctx, []storage.Article{article}); err != nil {
		logger.WithError(err).Error("could not store articles")
		response.MustRender(w, r, response.ErrUnknown(err))
		return
	}

	render.Status(r, http.StatusOK)
}

func (s *ArticleService) deleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.RequestLogger(r)
	slug := chi.URLParam(r, "slug")

	if err := s.store.DeleteArticles(ctx, []storage.Article{{Slug: slug}}); err != nil {
		logger.WithError(err).Error("could not delete articles")
		response.MustRender(w, r, response.ErrUnknown(err))
		return
	}

	render.NoContent(w, r)
}

func newArticleListResponse(articles []storage.Article) []render.Renderer {
	var list []render.Renderer
	for _, article := range articles {
		list = append(list, newArticleResponse(article))
	}
	return list
}

func newArticleResponse(article storage.Article) *articleResponse {
	return &articleResponse{
		ID:    article.ID,
		Title: article.Title,
		Slug:  article.Slug,
	}
}

type articleResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

func (*articleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type articleRequest struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

func (r *articleRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is empty")
	}
	if r.Slug == "" {
		return errors.New("slug is empty")
	}
	return nil
}
