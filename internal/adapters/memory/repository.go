package memory

import (
	"context"
	"sync"

	"github.com/grigory/url-shortener/internal/domain"
	"github.com/grigory/url-shortener/internal/ports"
)

var _ ports.Repository = (*Repository)(nil)

type Repository struct {
	mu          sync.RWMutex
	byShortCode map[string]domain.URL
	byOriginal  map[string]domain.URL
}

func New() *Repository {
	return &Repository{
		byShortCode: make(map[string]domain.URL),
		byOriginal:  make(map[string]domain.URL),
	}
}

func (r *Repository) Save(_ context.Context, u domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byOriginal[u.OriginalURL]; exists {
		return domain.ErrAlreadyExists
	}
	r.byShortCode[u.ShortCode] = u
	r.byOriginal[u.OriginalURL] = u
	return nil
}

func (r *Repository) FindByShortCode(_ context.Context, shortCode string) (domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byShortCode[shortCode]
	if !ok {
		return domain.URL{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *Repository) FindByOriginalURL(_ context.Context, originalURL string) (domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byOriginal[originalURL]
	if !ok {
		return domain.URL{}, domain.ErrNotFound
	}
	return u, nil
}
