package ports

import (
	"context"

	"github.com/grigory/url-shortener/internal/domain"
)

type Repository interface {
	Save(ctx context.Context, u domain.URL) error
	FindByShortCode(ctx context.Context, shortCode string) (domain.URL, error)
	FindByOriginalURL(ctx context.Context, originalURL string) (domain.URL, error)
}
