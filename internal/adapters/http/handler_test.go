package http_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	httpadapter "github.com/grigory/url-shortener/internal/adapters/http"
	"github.com/grigory/url-shortener/internal/domain"
	"github.com/grigory/url-shortener/internal/ports/mocks"
)

const (
	testOriginalURL = "https://example.com/some/long/path"
	testShortCode   = "abcdefghij"
)

func newTestRouter(t *testing.T) (http.Handler, *mocks.MockURLShortener) {
	t.Helper()
	svc := mocks.NewMockURLShortener(t)
	h := httpadapter.NewHandler(svc, slog.Default())
	return httpadapter.NewRouter(h), svc
}

func TestShortenURL_Success(t *testing.T) {
	router, svc := newTestRouter(t)

	svc.On("Shorten", mock.Anything, testOriginalURL).
		Return(domain.URL{ShortCode: testShortCode, OriginalURL: testOriginalURL}, nil)

	body := `{"url":"` + testOriginalURL + `"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, testShortCode, resp["short_code"])
	assert.Contains(t, resp["short_url"], "/"+testShortCode)
}

func TestShortenURL_InvalidBody(t *testing.T) {
	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShortenURL_InvalidURL(t *testing.T) {
	router, svc := newTestRouter(t)

	svc.On("Shorten", mock.Anything, "not-a-url").
		Return(domain.URL{}, domain.ErrInvalidURL)

	body := `{"url":"not-a-url"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestResolveURL_Success(t *testing.T) {
	router, svc := newTestRouter(t)

	svc.On("Resolve", mock.Anything, testShortCode).
		Return(domain.URL{ShortCode: testShortCode, OriginalURL: testOriginalURL}, nil)

	req := httptest.NewRequest(http.MethodGet, "/"+testShortCode, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, testOriginalURL, resp["url"])
}

func TestResolveURL_NotFound(t *testing.T) {
	router, svc := newTestRouter(t)

	svc.On("Resolve", mock.Anything, "xxxxxxxxxx").
		Return(domain.URL{}, domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/xxxxxxxxxx", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
