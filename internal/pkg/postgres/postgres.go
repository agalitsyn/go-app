package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/agalitsyn/go-app/internal/pkg/log"
)

type DB struct {
	cfg *pgxpool.Config

	Session *pgxpool.Pool
	Logger  log.Logger
}

func New(connString string, logger log.Logger) (*DB, error) {
	connCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("connection string parse error: %w", err)
	}

	// If you use pgpool, you need to use simple protocol
	//connCfg.ConnConfig.PreferSimpleProtocol = true
	//connCfg.ConnConfig.BuildStatementCache = func(conn *pgconn.PgConn) stmtcache.Cache {
	//	return stmtcache.New(conn, stmtcache.ModeDescribe, 512)
	//}

	return &DB{cfg: connCfg, Logger: logger}, nil
}

func (d *DB) Connect(ctx context.Context) error {
	var err error
	maxAttempts := 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var pool *pgxpool.Pool
		pool, err = pgxpool.ConnectConfig(ctx, d.cfg)
		if err == nil {
			d.Session = pool
			break
		}
		nextAttemptWait := time.Duration(attempt) * time.Second
		d.Logger.Warnf("attempt %v: could not establish a connection with the database, wait for %v.", attempt, nextAttemptWait)
		time.Sleep(nextAttemptWait)
	}
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}
	return nil
}
