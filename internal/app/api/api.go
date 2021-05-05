package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/goware/cors"

	"github.com/agalitsyn/go-app/internal/pkg/health"
	"github.com/agalitsyn/go-app/internal/pkg/log"
	mw "github.com/agalitsyn/go-app/internal/pkg/middleware"
	"github.com/agalitsyn/go-app/internal/pkg/response"
)

type Config struct {
	CORSOptions cors.Options
	DocsPath    string
}

func New(cfg Config, logger *log.StructuredLogger, articleService *ArticleService) chi.Router {
	r := chi.NewRouter()
	r.Use( // note: order of middlewares is important
		middleware.RequestID,
		middleware.RealIP,
		mw.RequestLogger(logger),
		middleware.Recoverer,
		cors.New(cfg.CORSOptions).Handler,
	)

	r.Mount("/readiness", health.Routes())
	r.Route("/1.0", func(r chi.Router) {
		r.Use(mw.APIVersion("1.0"))

		r.Mount("/articles", articleService.Routes())
	})

	response.FileServer(r, "/docs", http.Dir(cfg.DocsPath))

	return r
}
