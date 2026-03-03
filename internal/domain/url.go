package domain

import (
	"errors"
	"net/url"
	"time"
)

type URL struct {
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
}

var (
	ErrNotFound      = errors.New("url not found")
	ErrAlreadyExists = errors.New("url already exists")
	ErrInvalidURL    = errors.New("invalid url")
)

func Validate(rawURL string) error {
	if rawURL == "" {
		return ErrInvalidURL
	}
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ErrInvalidURL
	}
	return nil
}
