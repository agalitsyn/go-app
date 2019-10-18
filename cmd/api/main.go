package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/goware/cors"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/agalitsyn/goapi/article"
	"github.com/agalitsyn/goapi/cmd/internal/health"
	"github.com/agalitsyn/goapi/handler"
	"github.com/agalitsyn/goapi/log"
	"github.com/agalitsyn/goapi/postgres"

	"github.com/agalitsyn/goapi/cmd/internal/flag"
)

type CliFlags struct {
	DocsPath string `long:"docs-path" env:"API_DOCS_PATH" default:"docs" description:"Path to documentation folder."`

	HTTP struct {
		Addr           string   `long:"addr" env:"API_HTTP_ADDR" default:"localhost:5000" description:"HTTP service address."`
		AllowedOrigins []string `long:"allowed-origins" env:"API_ALLOWED_ORIGINS" description:"The list of origins a cross-domain request can be executed from."`
		AllowedHeaders []string `long:"allowed-headers" env:"API_ALLOWED_HEADERS" description:"The list of non simple headers the client is allowed to use with cross-domain requests."`
		ExposedHeaders []string `long:"exposed-headers" env:"API_EXPOSED_ORIGINS" description:"The list which indicates which headers are safe to expose."`
	}

	Postgres struct {
		URL                string        `long:"postgres-url" env:"API_POSTGRES_URL" default:"postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable" description:"URL to PostgreSQL database."`
		MaxConnLifetimeSec time.Duration `long:"postgres-conn-lt" env:"API_POSTGRES_MAX_CONN_LT" default:"60"`
		MaxIdleConns       int           `long:"postgres-max-idle-conns" env:"API_POSTGRES_MAX_IDLE_CONN" default:"1"`
		MaxOpenConns       int           `long:"postgres-max-open-conn" env:"API_POSTGRES_MAX_OPEN_CONN" default:"1"`
	}

	Log struct {
		Level  string `long:"log-level" default:"info" choice:"debug" choice:"info" choice:"warn" choice:"error" env:"API_LOG_LEVEL" description:"Log level."`
		Format string `long:"log-format" default:"text" choice:"text" choice:"json" env:"API_LOG_FORMAT" description:"Log format."`
	}

	PrintVersion bool `long:"version" description:"Show application version"`
}

func main() {
	var cfg CliFlags
	flag.ParseFlags(&cfg)

	logger := log.New(cfg.Log.Format, cfg.Log.Level, os.Stdout)
	logger.Infof("started with config: %+v", cfg)

	pcfg := postgres.Config{
		MaxConnLifetime: cfg.Postgres.MaxConnLifetimeSec,
		MaxOpenConns:    cfg.Postgres.MaxOpenConns,
		MaxIdleConns:    cfg.Postgres.MaxIdleConns,
	}
	db, err := initDatabase(cfg.Postgres.URL, logger, pcfg)
	if err != nil {
		logger.WithError(err).Fatal()
	}
	defer db.Close()
	articleManager := article.NewManager(db.DB)

	cm := cors.New(cors.Options{
		AllowedOrigins:   cfg.HTTP.AllowedOrigins,
		AllowedHeaders:   cfg.HTTP.AllowedHeaders,
		ExposedHeaders:   cfg.HTTP.ExposedHeaders,
		AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodDelete},
		AllowCredentials: true,
	})
	// note: order of middlewares is important
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		handler.RequestLogger(logger),
		middleware.Recoverer,
		cm.Handler,
	)
	r.Mount("/readiness", health.Routes())
	r.Route("/1.0", func(r chi.Router) {
		r.Use(handler.ApiVersion("1.0"))
		// TODO: add urls from packages here
		r.Mount("/articles", article.Routes(articleManager))
	})
	handler.FileServer(r, "/docs", http.Dir(cfg.DocsPath))
	srv := &http.Server{Addr: cfg.HTTP.Addr, Handler: r}

	sigquit := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(sigquit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigquit
		logger.Infof("captured %v, exiting...", s)

		health.SetReadinessStatus(http.StatusServiceUnavailable)

		logger.Info("gracefully shutdown server")
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.WithError(err).Error("could not shutdown server")
		}
	}()

	logger.Info("starting http service...")
	logger.Infof("listening on %s", cfg.HTTP.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Error("server error")
	}
}

func initDatabase(dsn string, logger log.Logger, pcfg postgres.Config) (*postgres.Database, error) {
	migrations := []*migrate.Migration{}
	// TODO: add migrations from packages here
	migrations = append(migrations, article.Migrations()...)
	ms := &migrate.MemoryMigrationSource{Migrations: migrations}

	db, err := postgres.New(dsn, logger, pcfg)
	if err != nil {
		return nil, err
	}
	if err := db.Connect(); err != nil {
		return nil, err
	}
	if err := db.Migrate(ms); err != nil {
		return nil, err
	}
	return db, nil
}
