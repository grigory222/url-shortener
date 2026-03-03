package memory_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grigory/url-shortener/internal/adapters/memory"
	"github.com/grigory/url-shortener/internal/domain"
)

func newURL(short, original string) domain.URL {
	return domain.URL{
		ShortCode:   short,
		OriginalURL: original,
		CreatedAt:   time.Now().UTC(),
	}
}

func TestSave_Success(t *testing.T) {
	r := memory.New()
	u := newURL("abcdefghij", "https://example.com")

	err := r.Save(context.Background(), u)
	require.NoError(t, err)
}

func TestSave_DuplicateOriginalURL(t *testing.T) {
	r := memory.New()
	u := newURL("abcdefghij", "https://example.com")
	require.NoError(t, r.Save(context.Background(), u))

	err := r.Save(context.Background(), newURL("xxxxxxxxxx", "https://example.com"))
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestFindByShortCode_Found(t *testing.T) {
	r := memory.New()
	u := newURL("abcdefghij", "https://example.com")
	require.NoError(t, r.Save(context.Background(), u))

	got, err := r.FindByShortCode(context.Background(), "abcdefghij")
	require.NoError(t, err)
	assert.Equal(t, u.OriginalURL, got.OriginalURL)
}

func TestFindByShortCode_NotFound(t *testing.T) {
	r := memory.New()

	_, err := r.FindByShortCode(context.Background(), "xxxxxxxxxx")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestFindByOriginalURL_Found(t *testing.T) {
	r := memory.New()
	u := newURL("abcdefghij", "https://example.com")
	require.NoError(t, r.Save(context.Background(), u))

	got, err := r.FindByOriginalURL(context.Background(), "https://example.com")
	require.NoError(t, err)
	assert.Equal(t, u.ShortCode, got.ShortCode)
}

func TestFindByOriginalURL_NotFound(t *testing.T) {
	r := memory.New()

	_, err := r.FindByOriginalURL(context.Background(), "https://example.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestConcurrentSave(t *testing.T) {
	const goroutines = 100
	r := memory.New()

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			u := newURL(
				randCode(idx),
				"https://example.com/"+randCode(idx),
			)
			_ = r.Save(context.Background(), u)
		}(i)
	}
	wg.Wait()
}

func randCode(idx int) string {
	const alpha = "abcdefghijklmnopqrstuvwxyz"
	return string([]byte{
		alpha[(idx)%26],
		alpha[(idx+1)%26],
		alpha[(idx+2)%26],
		alpha[(idx+3)%26],
		alpha[(idx+4)%26],
		alpha[(idx+5)%26],
		alpha[(idx+6)%26],
		alpha[(idx+7)%26],
		byte('0' + idx%10),
		byte('0' + (idx+1)%10),
	})
}
