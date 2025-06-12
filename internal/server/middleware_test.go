package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/mocks"
)

func TestLogHandler(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockShortener) {})
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	c := srv.echo.NewContext(req, rec)

	var called bool
	handler := srv.logHandler(func(c echo.Context) error {
		called = true
		return c.String(http.StatusOK, "pong")
	})

	err := handler(c)
	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "pong", rec.Body.String())
}

func TestCORSMiddleware(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockShortener) {})
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	c := srv.echo.NewContext(req, rec)

	handler := srv.CORSMiddleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	require.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
}

func TestTrustedSubnetMiddleware(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockShortener) {})

	t.Run("trusted subnet - allowed", func(t *testing.T) {
		srv.cfg.Server.FlagTrustedSubnet = "192.168.1.0/24"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderXRealIP, "192.168.1.1") // trusted IP

		rec := httptest.NewRecorder()
		c := srv.echo.NewContext(req, rec)

		handler := srv.TrustedSubnetMiddleware(func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		err := handler(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("untrusted IP - forbidden", func(t *testing.T) {
		srv.cfg.Server.FlagTrustedSubnet = "192.168.1.0/24"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderXRealIP, "10.0.0.1") // untrusted IP

		rec := httptest.NewRecorder()
		c := srv.echo.NewContext(req, rec)

		handler := srv.TrustedSubnetMiddleware(func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		err := handler(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("no trusted subnet configured - forbidden", func(t *testing.T) {
		srv.cfg.Server.FlagTrustedSubnet = "" // trusted subnet disabled → block all

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderXRealIP, "192.168.1.1") // any IP

		rec := httptest.NewRecorder()
		c := srv.echo.NewContext(req, rec)

		handler := srv.TrustedSubnetMiddleware(func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		err := handler(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, rec.Code)
	})
}
