package database

import (
	"database/sql"

	"github.com/agalitsyn/goapi/log"

	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
)

type Database struct {
	DB     *sql.DB
	Logger log.Logger
}

func New(dsn string, logger log.Logger) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "could not open database")
	}
	return &Database{
		DB:     db,
		Logger: logger,
	}, nil
}

func (d *Database) Connect() error {
	var err error
	maxAttempts := 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = d.DB.Ping()
		if err == nil {
			break
		}
		nextAttemptWait := time.Duration(attempt) * time.Second
		d.Logger.Warnf("attempt %v: could not establish a connection with the database, wait for %v.", attempt, nextAttemptWait)
		time.Sleep(nextAttemptWait)
	}
	if err != nil {
		return errors.Wrap(err, "could not connect to database")
	}
	return nil
}

func (d *Database) Close() error {
	if err := d.DB.Close(); err != nil {
		return errors.Wrap(err, "could not close database")
	}
	return nil
}

func (d *Database) Migrate(migrations *migrate.MemoryMigrationSource) error {
	migrate.SetTable("migrations")
	done, err := migrate.Exec(d.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "could not perform database migrations")
	}
	d.Logger.Infof("performed %d migrations", done)
	return nil
}
