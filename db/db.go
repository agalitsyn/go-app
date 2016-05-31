package db

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func New(dsn string) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
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
		//log.WithError(dbError).Error("Could not establish a connection with the database")
		time.Sleep(time.Duration(attempts) * time.Second)
	}

	if dbError != nil {
		return dbError
	}
	return nil
}

func (d *Database) Close() {
	err := d.db.Close()
	if err != nil {
		//log.Fatal(err)
	}
}
