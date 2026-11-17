// Package postgres implementa links.Store su PostgreSQL.
package postgres

import (
	"context"
	"errors"

	"github.com/alessiogarifo92/linkboard/internal/links"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema minimo, applicato all'avvio. In un progetto reale useresti migrazioni versionate
// (in Kubernetes: un Job o un initContainer, vedi settimana 5).
const Schema = `
CREATE TABLE IF NOT EXISTS links (
    code       TEXT PRIMARY KEY,
    url        TEXT NOT NULL,
    clicks     BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS links_created_at_idx ON links (created_at DESC);`

type Store struct{ Pool *pgxpool.Pool }

func Connect(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Store{Pool: pool}, nil
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, Schema)
	return err
}

func (s *Store) Ping(ctx context.Context) error { return s.Pool.Ping(ctx) }

func (s *Store) Create(ctx context.Context, code, rawURL string) (links.Link, error) {
	l := links.Link{Code: code, URL: rawURL}
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO links (code, url) VALUES ($1, $2) RETURNING created_at`, code, rawURL).Scan(&l.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return links.Link{}, links.ErrCodeTaken
	}
	return l, err
}

func (s *Store) Get(ctx context.Context, code string) (links.Link, error) {
	var l links.Link
	err := s.Pool.QueryRow(ctx,
		`SELECT code, url, clicks, created_at FROM links WHERE code = $1`, code).
		Scan(&l.Code, &l.URL, &l.Clicks, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return links.Link{}, links.ErrNotFound
	}
	return l, err
}

func (s *Store) AddClicks(ctx context.Context, code string, n int64) error {
	tag, err := s.Pool.Exec(ctx, `UPDATE links SET clicks = clicks + $2 WHERE code = $1`, code, n)
	if err == nil && tag.RowsAffected() == 0 {
		return links.ErrNotFound
	}
	return err
}

func (s *Store) Recent(ctx context.Context, limit int) ([]links.Link, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT code, url, clicks, created_at FROM links ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (links.Link, error) {
		var l links.Link
		err := r.Scan(&l.Code, &l.URL, &l.Clicks, &l.CreatedAt)
		return l, err
	})
}
