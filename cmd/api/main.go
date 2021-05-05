package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/goware/cors"

	"github.com/agalitsyn/go-app/internal/app/api"
	"github.com/agalitsyn/go-app/internal/pkg/flag"
	"github.com/agalitsyn/go-app/internal/pkg/health"
	"github.com/agalitsyn/go-app/internal/pkg/log"
	"github.com/agalitsyn/go-app/internal/pkg/postgres"
	"github.com/agalitsyn/go-app/internal/storage/rdb"
)

type CliFlags struct {
	DocsPath string `long:"docs-path" env:"DOCS_PATH" default:"docs" description:"Path to documentation folder."`

	HTTP struct {
		Addr           string   `long:"addr" env:"HTTP_ADDR" default:"localhost:8080" description:"HTTP service address."`
		AllowedOrigins []string `long:"allowed-origins" env:"ALLOWED_ORIGINS" description:"The list of origins a cross-domain request can be executed from."`
		AllowedHeaders []string `long:"allowed-headers" env:"ALLOWED_HEADERS" description:"The list of non simple headers the client is allowed to use with cross-domain requests."`
		ExposedHeaders []string `long:"exposed-headers" env:"EXPOSED_ORIGINS" description:"The list which indicates which headers are safe to expose."`
	}

	Postgres struct {
		URL string `long:"postgres-url" env:"POSTGRES_URL" default:"postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable" description:"URL to PostgreSQL database."`
	}

	//nolint[:staticcheck]
	Log struct {
		Level  string `long:"log-level" default:"info" choice:"debug" choice:"info" choice:"warn" choice:"error" env:"LOG_LEVEL" description:"Log level."`
		Format string `long:"log-format" default:"text" choice:"text" choice:"json" env:"LOG_FORMAT" description:"Log format."`
	}

	PrintVersion bool `long:"version" description:"Show application version"`
}

func main() {
	var cfg CliFlags
	flag.ParseFlags(&cfg)

	logger := log.New(cfg.Log.Format, cfg.Log.Level, os.Stdout)
	logger.Debugf("started with config: %+v", cfg)

	ctx := context.Background()

	pg, err := postgres.New(cfg.Postgres.URL, logger)
	if err != nil {
		logger.Fatalf("could not init postgres: %s", err)
	}
	defer pg.Session.Close()

	if err = pg.Connect(ctx); err != nil {
		logger.WithError(err).Error("could not connect to postgres")
	}

	articleStorage := rdb.NewArticleStorage(pg)

	apiCfg := api.Config{
		CORSOptions: cors.Options{
			AllowedOrigins:   cfg.HTTP.AllowedOrigins,
			AllowedHeaders:   cfg.HTTP.AllowedHeaders,
			ExposedHeaders:   cfg.HTTP.ExposedHeaders,
			AllowedMethods:   []string{http.MethodGet, http.MethodPut, http.MethodDelete},
			AllowCredentials: true,
		},
		DocsPath: cfg.DocsPath,
	}
	r := api.New(apiCfg, logger, api.NewArticleService(articleStorage))

	srv := &http.Server{Addr: cfg.HTTP.Addr, Handler: r}

	sigquit := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(sigquit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigquit
		logger.Infof("captured %v, exiting...", s)

		health.SetReadinessStatus(http.StatusServiceUnavailable)

		logger.Info("gracefully shutdown server")
		if err = srv.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("could not shutdown server")
		}
	}()

	logger.Info("starting http service...")
	logger.Infof("listening on %s", cfg.HTTP.Addr)
	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Error("server error")
	}
}
