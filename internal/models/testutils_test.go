package models

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open(
		"mysql",
		"test_web:pass@/test_snippetbox?parseTime=true&multiStatements=true",
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	setupSQL, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		_ = db.Close()
		t.Fatal(err)
	}

	if _, err = db.ExecContext(ctx, string(setupSQL)); err != nil {
		_ = db.Close()
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

		if _, err = db.ExecContext(ctx, string(teardownSQL)); err != nil {
			t.Fatal(err)
		}

		_ = db.Close()
	})

	return db
}
