package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/mocks"
)

func TestAuthMiddleware_NewCookie(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	ctx := e.NewContext(req, rec)

	called := false
	next := func(c echo.Context) error {
		called = true
		val := c.Get("userID")
		require.IsType(t, int(0), val)
		require.NotZero(t, val)
		return c.String(http.StatusOK, "ok")
	}

	err := srv.authMiddleware(next)(ctx)
	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, http.StatusOK, rec.Code)

	resp := rec.Result()
	defer resp.Body.Close()

	cookies := resp.Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	require.Equal(t, cookieName, cookie.Name)
	require.NotEmpty(t, cookie.Value)
}

func TestAuthMiddleware_ValidCookie(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	token, err := srv.createJWT()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
	})

	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	var captured any
	next := func(c echo.Context) error {
		captured = c.Get("userID")
		return c.String(http.StatusOK, "ok")
	}

	err = srv.authMiddleware(next)(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.IsType(t, int(0), captured)
	require.NotZero(t, captured)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:     cookieName,
		Value:    "bad.token.value",
		HttpOnly: true,
	})

	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	next := func(c echo.Context) error {
		t.Fatal("next should not be called")
		return nil
	}

	err := srv.authMiddleware(next)(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Getting user from cookie")
}
