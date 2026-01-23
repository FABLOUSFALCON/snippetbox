// Package models contains data models and database interaction logic.
package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SnippetModelInterface interface {
	Insert(ctx context.Context, title, content string, expires int) (int, error)
	Get(ctx context.Context, id int) (Snippet, error)
	Latest(ctx context.Context) ([]Snippet, error)
}

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *pgxpool.Pool
}

func (m *SnippetModel) Insert(ctx context.Context, title, content string, expires int) (int, error) {
	stmt := `
		INSERT INTO snippets (title, content, created, expires)
		VALUES ($1, $2, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC' + $3 * INTERVAL '1 day')
		RETURNING id
	`

	var id int
	err := m.DB.QueryRow(ctx, stmt, title, content, expires).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *SnippetModel) Get(ctx context.Context, id int) (Snippet, error) {
	stmt := `
		SELECT id, title, content, created, expires
		FROM snippets
		WHERE expires > NOW() AT TIME ZONE 'UTC' AND id = $1
	`

	row := m.DB.QueryRow(ctx, stmt, id)

	var s Snippet
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		}
		return Snippet{}, err
	}

	return s, nil
}

func (m *SnippetModel) Latest(ctx context.Context) ([]Snippet, error) {
	stmt := `
		SELECT id, title, content, created, expires
		FROM snippets
		WHERE expires > NOW() AT TIME ZONE 'UTC'
		ORDER BY id DESC
		LIMIT 10
	`

	rows, err := m.DB.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snippets []Snippet

	for rows.Next() {
		var s Snippet
		err := rows.Scan(
			&s.ID,
			&s.Title,
			&s.Content,
			&s.Created,
			&s.Expires,
		)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
