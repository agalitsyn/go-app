package main

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gravitational/trace"

	_ "github.com/lib/pq"

	log "github.com/Sirupsen/logrus"
)

var (
	ErrCantParseConfig      = errors.New("Can't parse config")
	ErrDBConnAttemptsFailed = errors.New("All attempts failed")
)

type Database struct {
	*sql.DB
}

func GetDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &Database{db}, nil
}

func (d *Database) Connect() error {
	var dbError error

	maxAttempts := 30
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		dbError = d.Ping()
		if dbError == nil {
			break
		}
		nextAttemptWait := time.Duration(attempt) * time.Second

		log.Errorf("Attempt %v: could not establish a connection with the database. Wait for %v.", attempt, nextAttemptWait)
		time.Sleep(nextAttemptWait)
	}

	if dbError != nil {
		return trace.Wrap(ErrDBConnAttemptsFailed)
	}
	return nil
}

func (d *Database) Close() error {
	if err := d.Close(); err != nil {
		return trace.Wrap(err, "Can't close database")
	}
	return nil
}
