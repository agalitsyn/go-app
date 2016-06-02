package db

import (
	"database/sql"
	"time"

	"github.com/agalitsyn/goapi/log"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var (
	DBDriverName            = "postgres"
	ErrDBConnAttemptsFailed = errors.New("All attempts failed")
)

type Database struct {
	db  *sql.DB
	log log.Logger
}

func New(dsn string) (*Database, error) {
	fields := map[string]interface{}{
		"driver": DBDriverName,
	}
	dbLogger := log.GetLoggerWithFields("db", fields)

	db, err := sql.Open(DBDriverName, dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Can't open database")
	}

	return &Database{
		db:  db,
		log: dbLogger,
	}, nil
}

func (d *Database) Connect() error {
	var dbError error

	maxAttempts := 30
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		dbError = d.db.Ping()
		if dbError == nil {
			break
		}
		nextAttemptWait := time.Duration(attempt) * time.Second

		d.log.WithError(dbError).Errorf("Attempt %v: could not establish a connection with the database. Wait for %v.", attempt, nextAttemptWait)
		time.Sleep(nextAttemptWait)
	}

	if dbError != nil {
		return ErrDBConnAttemptsFailed
	}
	return nil
}

func (d *Database) Close() error {
	err := d.db.Close()
	if err != nil {
		return errors.Wrap(err, "Can't close database")
	}
	return nil
}
