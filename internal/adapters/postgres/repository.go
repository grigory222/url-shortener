package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grigory/url-shortener/internal/domain"
	"github.com/grigory/url-shortener/internal/ports"
)

var _ ports.Repository = (*Repository)(nil)

const pgUniqueViolation = "23505"

type Repository struct {
	q *Queries
}

func NewRepository(ctx context.Context, dsn string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: open pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return &Repository{q: New(pool)}, nil
}

func (r *Repository) Save(ctx context.Context, u domain.URL) error {
	_, err := r.q.InsertURL(ctx, InsertURLParams{
		ShortCode:   u.ShortCode,
		OriginalUrl: u.OriginalURL,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("postgres: insert url: %w", err)
	}
	return nil
}

func (r *Repository) FindByShortCode(ctx context.Context, shortCode string) (domain.URL, error) {
	row, err := r.q.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.URL{}, domain.ErrNotFound
		}
		return domain.URL{}, fmt.Errorf("postgres: find by short code: %w", err)
	}
	return toDomain(row), nil
}

func (r *Repository) FindByOriginalURL(ctx context.Context, originalURL string) (domain.URL, error) {
	row, err := r.q.GetURLByOriginalURL(ctx, originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.URL{}, domain.ErrNotFound
		}
		return domain.URL{}, fmt.Errorf("postgres: find by original url: %w", err)
	}
	return toDomain(row), nil
}

func toDomain(u Url) domain.URL {
	return domain.URL{
		ShortCode:   u.ShortCode,
		OriginalURL: u.OriginalUrl,
		CreatedAt:   u.CreatedAt.Time,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
