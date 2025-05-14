package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/mocks"
)

func TestGzipMiddleware_SetsResponseHeader(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	c := echo.New().NewContext(req, rec)
	c.Response().Header().Set("Content-Type", "application/json")

	handler := srv.gzipMiddleware(func(c echo.Context) error {
		_, err := c.Response().Write([]byte("hello"))
		return err
	})

	err := handler(c)
	require.NoError(t, err)

	// Проверим, что Content-Encoding проставлен
	require.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
}

func TestGzipMiddleware_DecodesGzippedRequest(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(`{"test":1}`))
	_ = gz.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	handler := srv.gzipMiddleware(func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"test":1}`, string(body))
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}
