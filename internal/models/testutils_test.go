package models

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Use PostgreSQL test database
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://test_web:pass@localhost/test_snippetbox?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	if err = db.Ping(ctx); err != nil {
		db.Close()
		t.Fatal(err)
	}

	setupSQL, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	if _, err = db.Exec(ctx, string(setupSQL)); err != nil {
		db.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		//nolint:govet // shadow ctx variable is no issue here
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		teardownSQL, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}

		if _, err = db.Exec(ctx, string(teardownSQL)); err != nil {
			t.Fatal(err)
		}

		db.Close()
	})

	return db
}
