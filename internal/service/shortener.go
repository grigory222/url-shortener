package service

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"time"

	"github.com/grigory/url-shortener/internal/domain"
	"github.com/grigory/url-shortener/internal/ports"
)

const (
	shortCodeLen = 10
	alphabet     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	maxRetries   = 10
)

var _ ports.URLShortener = (*Shortener)(nil)

type Shortener struct {
	repo   ports.Repository
	logger *slog.Logger
}

func New(repo ports.Repository, logger *slog.Logger) *Shortener {
	return &Shortener{repo: repo, logger: logger}
}

func (s *Shortener) Shorten(ctx context.Context, originalURL string) (domain.URL, error) {
	if err := domain.Validate(originalURL); err != nil {
		return domain.URL{}, err
	}

	existing, err := s.repo.FindByOriginalURL(ctx, originalURL)
	if err == nil {
		s.logger.InfoContext(ctx, "existing short code returned",
			slog.String("short_code", existing.ShortCode),
			slog.String("original_url", originalURL),
		)
		return existing, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.URL{}, err
	}

	for range maxRetries {
		code, genErr := generateShortCode()
		if genErr != nil {
			return domain.URL{}, genErr
		}

		u := domain.URL{
			ShortCode:   code,
			OriginalURL: originalURL,
			CreatedAt:   time.Now().UTC(),
		}

		saveErr := s.repo.Save(ctx, u)
		if saveErr == nil {
			s.logger.InfoContext(ctx, "short code created",
				slog.String("short_code", code),
				slog.String("original_url", originalURL),
			)
			return u, nil
		}

		if errors.Is(saveErr, domain.ErrAlreadyExists) {
			found, findErr := s.repo.FindByOriginalURL(ctx, originalURL)
			if findErr == nil {
				return found, nil
			}
			return domain.URL{}, findErr
		}

		s.logger.WarnContext(ctx, "short code collision, retrying",
			slog.String("code", code),
			slog.String("error", saveErr.Error()),
		)
	}

	return domain.URL{}, errors.New("service: failed to generate a unique short code")
}

func (s *Shortener) Resolve(ctx context.Context, shortCode string) (domain.URL, error) {
	u, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.InfoContext(ctx, "short code not found",
				slog.String("short_code", shortCode),
			)
		}
		return domain.URL{}, err
	}
	return u, nil
}

func generateShortCode() (string, error) {
	const alphabetLen = byte(len(alphabet))
	const limit = (256 / int(alphabetLen)) * int(alphabetLen)

	result := make([]byte, shortCodeLen)
	buf := make([]byte, shortCodeLen*2)

	filled := 0
	for filled < shortCodeLen {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if filled == shortCodeLen {
				break
			}
			if int(b) < limit {
				result[filled] = alphabet[int(b)%int(alphabetLen)]
				filled++
			}
		}
	}
	return string(result), nil
}
