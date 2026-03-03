package service_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/grigory/url-shortener/internal/domain"
	"github.com/grigory/url-shortener/internal/ports/mocks"
	"github.com/grigory/url-shortener/internal/service"
)

const testURL = "https://example.com/some/very/long/path"

func newService(t *testing.T) (*service.Shortener, *mocks.MockRepository) {
	t.Helper()
	repo := mocks.NewMockRepository(t)
	svc := service.New(repo, slog.Default())
	return svc, repo
}

func TestShorten_NewURL(t *testing.T) {
	ctx := context.Background()
	svc, repo := newService(t)

	repo.On("FindByOriginalURL", ctx, testURL).Return(domain.URL{}, domain.ErrNotFound)
	repo.On("Save", ctx, matchAnyURL(testURL)).Return(nil)

	u, err := svc.Shorten(ctx, testURL)

	require.NoError(t, err)
	assert.Equal(t, testURL, u.OriginalURL)
	assert.Len(t, u.ShortCode, 10)
	assert.Regexp(t, `^[a-zA-Z0-9_]{10}$`, u.ShortCode)
}

func TestShorten_ExistingURL(t *testing.T) {
	ctx := context.Background()
	svc, repo := newService(t)

	existing := domain.URL{ShortCode: "abcdefghij", OriginalURL: testURL}
	repo.On("FindByOriginalURL", ctx, testURL).Return(existing, nil)

	u, err := svc.Shorten(ctx, testURL)

	require.NoError(t, err)
	assert.Equal(t, existing.ShortCode, u.ShortCode)
}

func TestShorten_InvalidURL(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	_, err := svc.Shorten(ctx, "not-a-url")
	assert.ErrorIs(t, err, domain.ErrInvalidURL)
}

func TestShorten_EmptyURL(t *testing.T) {
	ctx := context.Background()
	svc, _ := newService(t)

	_, err := svc.Shorten(ctx, "")
	assert.ErrorIs(t, err, domain.ErrInvalidURL)
}

func TestShorten_RaceCondition(t *testing.T) {
	ctx := context.Background()
	svc, repo := newService(t)

	existing := domain.URL{ShortCode: "aaaaaaaaaa", OriginalURL: testURL}
	repo.On("FindByOriginalURL", ctx, testURL).Return(domain.URL{}, domain.ErrNotFound).Once()
	repo.On("Save", ctx, matchAnyURL(testURL)).Return(domain.ErrAlreadyExists)
	repo.On("FindByOriginalURL", ctx, testURL).Return(existing, nil)

	u, err := svc.Shorten(ctx, testURL)
	require.NoError(t, err)
	assert.Equal(t, existing.ShortCode, u.ShortCode)
}

func TestResolve_Found(t *testing.T) {
	ctx := context.Background()
	svc, repo := newService(t)

	expected := domain.URL{ShortCode: "abcdefghij", OriginalURL: testURL}
	repo.On("FindByShortCode", ctx, "abcdefghij").Return(expected, nil)

	u, err := svc.Resolve(ctx, "abcdefghij")
	require.NoError(t, err)
	assert.Equal(t, testURL, u.OriginalURL)
}

func TestResolve_NotFound(t *testing.T) {
	ctx := context.Background()
	svc, repo := newService(t)

	repo.On("FindByShortCode", ctx, "xxxxxxxxxx").Return(domain.URL{}, domain.ErrNotFound)

	_, err := svc.Resolve(ctx, "xxxxxxxxxx")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func matchAnyURL(originalURL string) interface{} {
	return tmock.MatchedBy(func(u domain.URL) bool {
		return u.OriginalURL == originalURL
	})
}
