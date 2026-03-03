package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/grigory/url-shortener/internal/domain"
	"github.com/grigory/url-shortener/internal/ports"
)

type Handler struct {
	svc    ports.URLShortener
	logger *slog.Logger
}

func NewHandler(svc ports.URLShortener, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(newStructuredLogger(h.logger))
	r.Use(middleware.Recoverer)

	r.Post("/shorten", h.ShortenURL)
	r.Get("/{shortCode}", h.ResolveURL)

	return r
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.svc.Shorten(r.Context(), req.URL)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidURL):
			writeError(w, http.StatusUnprocessableEntity, "invalid url")
		default:
			h.logger.ErrorContext(r.Context(), "shorten error", slog.String("err", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, shortenResponse{
		ShortCode: u.ShortCode,
		ShortURL:  baseURL(r) + "/" + u.ShortCode,
	})
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + r.Host
}

type resolveResponse struct {
	URL string `json:"url"`
}

func (h *Handler) ResolveURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	u, err := h.svc.Resolve(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "short code not found")
			return
		}
		h.logger.ErrorContext(r.Context(), "resolve error", slog.String("err", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, resolveResponse{URL: u.OriginalURL})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func newStructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.InfoContext(r.Context(), "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
		})
	}
}
