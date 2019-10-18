package article

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/agalitsyn/goapi/handler"
	"github.com/agalitsyn/goapi/log"
)

func Routes(m *Manager) chi.Router {
	r := chi.NewRouter()

	r.Get("/", makeHandler(m, listHandler))

	r.Route("/{id}", func(r chi.Router) {
		r.Put("/", makeHandler(m, putHandler))
		r.Delete("/", makeHandler(m, deleteHandler))
	})

	return r
}

type handlerFunc func(m *Manager, w http.ResponseWriter, r *http.Request)

func makeHandler(m *Manager, handler handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(m, w, r)
	}
}

func listHandler(m *Manager, w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogEntry(r).WithField("context", "article")

	articles, err := m.All()
	if err != nil {
		logger.WithError(err).Error()
		render.Render(w, r, handler.ErrUnknown(err))
		return
	}
	if err := render.RenderList(w, r, newArticleListResponse(articles)); err != nil {
		logger.WithError(err).Error()
		render.Render(w, r, handler.ErrUnknown(err))
		return
	}
}

func putHandler(m *Manager, w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogEntry(r).WithField("context", "article")

	var data articleRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.WithError(err).Warn()
		render.Render(w, r, handler.ErrBadRequest(err))
		return
	}

	articleID := chi.URLParam(r, "id")
	article, err := m.ByID(articleID)
	if err != nil && err != ErrNotFound {
		logger.WithError(err).Error()
		render.Render(w, r, handler.ErrUnknown(err))
		return
	}
	if article == nil {
		d := &Article{
			ID:    articleID,
			Title: data.Title,
			Slug:  data.Slug,
		}
		if err := m.Save(d); err != nil {
			logger.WithError(err).Error()
			render.Render(w, r, handler.ErrUnknown(err))
			return
		}

		render.Status(r, http.StatusCreated)
		render.Render(w, r, newArticleResponse(d))
	} else {
		article.Title = data.Title
		article.Slug = data.Slug
		if err := m.Update(article); err != nil {
			logger.WithError(err).Error()
			render.Render(w, r, handler.ErrUnknown(err))
			return
		}

		render.Render(w, r, newArticleResponse(article))
	}
}

func deleteHandler(m *Manager, w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogEntry(r).WithField("context", "article")

	articleID := chi.URLParam(r, "id")
	article, err := m.ByID(articleID)
	if err != nil {
		if err == ErrNotFound {
			logger.WithError(err).Warn()
			render.Render(w, r, handler.ErrNotFound(err))
			return
		}
		logger.WithError(err).Error()
		render.Render(w, r, handler.ErrUnknown(err))
		return
	}
	if err := m.Delete(article); err != nil {
		logger.WithError(err).Error()
		render.Render(w, r, handler.ErrUnknown(err))
		return
	}
	render.NoContent(w, r)
}

func newArticleListResponse(articles []*Article) []render.Renderer {
	list := []render.Renderer{}
	for _, a := range articles {
		list = append(list, newArticleResponse(a))
	}
	return list
}

func newArticleResponse(article *Article) *articleResponse {
	return &articleResponse{article}
}

type articleResponse struct {
	*Article
}

func (dr *articleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type articleRequest struct {
	Title string `json:"title"`
	Slug  string `json:"slug"`
}
