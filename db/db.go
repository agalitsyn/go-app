package db

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Database struct {
	db *sql.DB
}

func New(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Can't open database")
	}

	return &Database{db}, nil
}

func (d *Database) Connect() error {
	var dbError error

	maxAttempts := 30
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		dbError = d.db.Ping()
		if dbError == nil {
			break
		}
		errors.Fprint(os.Stdout, errors.Wrap(dbError, "Could not establish a connection with the database"))
		time.Sleep(time.Duration(attempts) * time.Second)
	}

	if dbError != nil {
		return errors.Wrap(dbError, "All attempts failed")
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
