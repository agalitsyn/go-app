package rdb

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/agalitsyn/go-app/internal/pkg/log"
	"github.com/agalitsyn/go-app/internal/pkg/postgres"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*postgres.DB, func()) {
	return setupTestDBWithName(t, t.Name())
}

var testDBIdx uint32

const managementDB = "postgres"

func setupTestDBWithName(t *testing.T, name string) (*postgres.DB, func()) {
	t.Helper()
	if testing.Short() {
		t.Skip("skip long-running test in short mode")
	}

	ctx := context.Background()
	prefix := strings.ToLower(strings.Replace(name, "/", "_", -1))
	idx := atomic.AddUint32(&testDBIdx, 1)
	testDB := strings.ToLower(fmt.Sprintf("test_%d_%s_%d", time.Now().Unix(), prefix, idx))

	connString, ok := os.LookupEnv("POSTGRES_URL")
	if !ok {
		panic("POSTGRES_URL is required")
	}

	connCfg, err := pgxpool.ParseConfig(connString)
	require.NoError(t, err)
	connCfg.ConnConfig.Database = testDB

	managementConnCfg := connCfg.Copy()
	err = dbCreate(ctx, managementConnCfg.ConnConfig, testDB)
	require.NoError(t, err)

	conn, err := pgxpool.ConnectConfig(ctx, connCfg)
	require.NoError(t, err)

	teardown := func() {
		conn.Close()

		err = dbDrop(ctx, managementConnCfg.ConnConfig, testDB)
		require.NoError(t, err)
	}

	logger := log.New("", "", ioutil.Discard)
	return &postgres.DB{Session: conn, Logger: logger}, teardown
}

func dbCreate(ctx context.Context, cfg *pgx.ConnConfig, name string) error {
	cfg.Database = managementDB
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	if _, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", name)); err != nil {
		return err
	}
	if err = conn.Close(ctx); err != nil {
		return err
	}
	return nil
}

func dbDrop(ctx context.Context, cfg *pgx.ConnConfig, name string) error {
	cfg.Database = managementDB
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE %s", name))
	if err != nil {
		return err
	}
	if err = conn.Close(ctx); err != nil {
		return err
	}
	return nil
}

//nolint[:unused]
func countRows(ctx context.Context, table string, db *postgres.DB) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err := db.Session.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
