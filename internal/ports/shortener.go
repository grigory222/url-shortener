package ports

import (
	"context"

	"github.com/grigory/url-shortener/internal/domain"
)

type URLShortener interface {
	Shorten(ctx context.Context, originalURL string) (domain.URL, error)
	Resolve(ctx context.Context, shortCode string) (domain.URL, error)
}
