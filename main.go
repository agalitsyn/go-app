package main

import (
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pressly/chi"
	migrate "github.com/rubenv/sql-migrate"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/agalitsyn/goapi/article"
	"github.com/agalitsyn/goapi/database"
	"github.com/agalitsyn/goapi/health"
	"github.com/agalitsyn/goapi/log"
	"github.com/agalitsyn/goapi/middleware"
	"github.com/agalitsyn/goapi/router"
	"github.com/agalitsyn/goapi/service"
)

func main() {
	cfg := parseFlags()
	logger := log.New(cfg.log.format, cfg.log.level, os.Stdout)
	logger.Infof("started with config: %+v", cfg)

	logger.Info("connecting to the database")
	db, err := initDatabase(cfg.db.dsn, logger)
	if err != nil {
		logger.WithError(err).Fatal()
	}

	appRoutes, err := makeRoutes(db.DB, cfg.docsPath)
	if err != nil {
		logger.WithError(err).Fatal()
	}

	// note: order of middlewares is important
	r := router.New(
		router.WithRequestID(),
		router.WithRealIP(),
		router.WithLogging(logger),
		router.WithRecover(),
		router.WithCORS(cfg.http.allowedOrigins, cfg.http.allowedHeaders, cfg.http.exposedHeaders),
		appRoutes,
	)

	addr := net.JoinHostPort("", cfg.http.port)
	srv := service.New(addr, r)

	signalChan := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-signalChan
		logger.Infof("captured %v, exiting", s)
		health.SetReadinessStatus(http.StatusServiceUnavailable)

		db.Logger.Info("disconnecting from the database")
		db.Close()

		logger.Info("shutdown server")
		srv.Stop()

		switch s {
		case syscall.SIGINT:
			os.Exit(130)
		case syscall.SIGTERM:
			os.Exit(0)
		}
	}()

	logger.Info("starting http service...")
	logger.Infof("listening on %s", addr)
	if err = srv.Start(); err != nil {
		logger.WithError(err).Fatal()
	}
}

func initDatabase(dsn string, logger log.Logger) (*database.Database, error) {
	db, err := database.New(dsn, logger)
	if err != nil {
		return nil, err
	}
	if err := db.Connect(); err != nil {
		return nil, err
	}
	if err := db.Migrate(makeMigrations()); err != nil {
		return nil, err
	}
	return db, nil
}

// makeMigrations holds application-level migrations list
func makeMigrations() *migrate.MemoryMigrationSource {
	var migrations []*migrate.Migration

	migrations = append(migrations, article.Migrations()...)

	return &migrate.MemoryMigrationSource{
		Migrations: migrations,
	}
}

// makeRoutes holds application-level routes
func makeRoutes(db *sql.DB, docsDir string) (router.Option, error) {
	articleManager := article.NewManager(db)

	return func(r *router.Router) {
		r.FileServer("/docs", http.Dir(docsDir))
		r.Mount("/readiness", health.Routes())

		r.Route("/1.0", func(r chi.Router) {
			r.Use(middleware.ApiVersion("1.0"))

			r.Mount("/articles", article.Routes(articleManager))
		})
	}, nil
}

// cliFlags is a union of the fields, which applicaton could parse from CLI args
type cliFlags struct {
	docsPath string

	log struct {
		level  string
		format string
	}

	http struct {
		port           string
		allowedOrigins []string
		allowedHeaders []string
		exposedHeaders []string
	}

	db struct {
		dsn string
	}
}

// parseFlags maps CLI flags to struct
func parseFlags() *cliFlags {
	var cfg cliFlags

	kingpin.Flag("docs-path", "Path to documentation folder.").
		Default("docs").
		Envar("DOCS_PATH").
		StringVar(&cfg.docsPath)

	kingpin.Flag("log-level", "Log level.").
		Default("info").
		Envar("LOG_LEVEL").
		EnumVar(&cfg.log.level, "debug", "info", "warning", "error", "fatal", "panic")
	kingpin.Flag("log-format", "Log format.").
		Default("text").
		Envar("LOG_FORMAT").
		EnumVar(&cfg.log.format, "text", "json")

	kingpin.Flag("port", "HTTP service port.").
		Default("5000").
		Envar("PORT").
		StringVar(&cfg.http.port)
	kingpin.Flag("allowed-origins", "The list of origins a cross-domain request can be executed from.").
		Envar("ALLOWED_ORIGINS").
		PlaceHolder("domain").
		StringsVar(&cfg.http.allowedOrigins)
	kingpin.Flag("allowed-headers", "The list of non simple headers the client is allowed to use with cross-domain requests.").
		Envar("ALLOWED_HEADERS").
		PlaceHolder("header").
		StringsVar(&cfg.http.allowedHeaders)
	kingpin.Flag("exposed-headers", "The list which indicates which headers are safe to expose.").
		Envar("EXPOSED_HEADERS").
		PlaceHolder("domain").
		StringsVar(&cfg.http.exposedHeaders)

	kingpin.Flag("database-url", "URL to Postgresql 9.4 database.").
		Default("postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable&binary_parameters=yes").
		Envar("DATABASE_URL").
		StringVar(&cfg.db.dsn)

	kingpin.Parse()
	return &cfg
}
