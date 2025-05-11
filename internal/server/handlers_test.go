package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/mocks"
	"github.com/wickedv43/go-shortener/internal/storage"
)

func setupTestServer(tb testing.TB, configureMock func(*mocks.MockDataKeeper)) *Server {
	tb.Helper()

	ctrl := gomock.NewController(tb)

	container := do.New()

	do.Provide(container, func(i do.Injector) (*config.Config, error) {
		return &config.Config{Server: config.Server{FlagRunAddr: ":8080"}}, nil
	})
	do.Provide(container, logger.NewLogger)

	mockKeeper := mocks.NewMockDataKeeper(ctrl)
	configureMock(mockKeeper)

	do.Provide(container, func(i do.Injector) (storage.DataKeeper, error) {
		return mockKeeper, nil
	})

	do.Provide(container, NewServer)
	return do.MustInvoke[*Server](container)
}

func TestServer_create_success(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "https://example.com").
			Return(storage.Data{}, sql.ErrNoRows)

		mock.EXPECT().
			Save(gomock.Any(), gomock.AssignableToTypeOf(storage.Data{})).
			Return(nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 123)

	err := srv.Create(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "text/plain", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), "/")
}

func TestServer_create_invalidContentType(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"url":"https://example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 123)

	err := srv.Create(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Bad request")
}

func TestServer_create_unauthorized(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", "not-int")

	err := srv.Create(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "userID is not of type int")
}

func TestServer_create_conflict(t *testing.T) {
	existing := storage.Data{
		OriginalURL: "https://example.com",
		ShortURL:    "abc123",
		UUID:        123,
	}

	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "https://example.com").
			Return(existing, nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 123)

	err := srv.Create(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, rec.Body.String(), "abc123")
}

func TestServer_create_saveError(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "https://example.com").
			Return(storage.Data{}, sql.ErrNoRows)

		mock.EXPECT().
			Save(gomock.Any(), gomock.Any()).
			Return(errors.New("save error"))
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 123)

	err := srv.Create(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestServer_createJSON(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "https://example.com").
			Return(storage.Data{}, sql.ErrNoRows)

		mock.EXPECT().
			Save(gomock.Any(), gomock.AssignableToTypeOf(storage.Data{})).
			Return(nil)
	})

	reqBody := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 123)

	err := srv.CreateJSON(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), "/") // проверяем, что есть ShortURL
}

func TestServer_createJSON_conflict(t *testing.T) {
	existing := storage.Data{
		OriginalURL: "https://example.com",
		ShortURL:    "abc123",
		UUID:        123,
	}

	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "https://example.com").
			Return(existing, nil)
	})

	reqBody := `{"url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 123)

	err := srv.CreateJSON(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, rec.Code)
	require.Contains(t, rec.Body.String(), "abc123") // должен вернуть существующий short
}

func TestServer_getShort_ok(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "abc123").
			Return(storage.Data{
				OriginalURL: "https://example.com",
				ShortURL:    "abc123",
				DeletedFlag: false,
			}, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.SetParamNames("short")
	ctx.SetParamValues("abc123")

	err := srv.GetShort(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	require.Equal(t, "https://example.com", rec.Header().Get("Location"))
}

func TestServer_getShort_gone(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "abc123").
			Return(storage.Data{
				OriginalURL: "https://example.com",
				ShortURL:    "abc123",
				DeletedFlag: true,
			}, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.SetParamNames("short")
	ctx.SetParamValues("abc123")

	err := srv.GetShort(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusGone, rec.Code)
	require.Contains(t, rec.Body.String(), "Gone")
}

func TestServer_getShort_internalError(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "abc123").
			Return(storage.Data{}, errors.New("db down"))
	})

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.SetParamNames("short")
	ctx.SetParamValues("abc123")

	err := srv.GetShort(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Server error")
}

func TestServer_batch_success(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "https://a.com").
			Return(storage.Data{}, sql.ErrNoRows)

		mock.EXPECT().
			Save(gomock.Any(), gomock.AssignableToTypeOf(storage.Data{})).
			Return(nil)

		mock.EXPECT().
			Get(gomock.Any(), "https://b.com").
			Return(storage.Data{}, sql.ErrNoRows)

		mock.EXPECT().
			Save(gomock.Any(), gomock.AssignableToTypeOf(storage.Data{})).
			Return(nil)
	})

	reqBody := `[{"correlation_id": "1", "original_url": "https://a.com"},
	             {"correlation_id": "2", "original_url": "https://b.com"}]`

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/Batch", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 42)

	err := srv.Batch(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Contains(t, rec.Body.String(), `"correlation_id":"1"`)
	require.Contains(t, rec.Body.String(), `"correlation_id":"2"`)
}

func TestServer_batch_unauthorized(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/Batch", strings.NewReader(`[]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", "string-instead-of-int")

	err := srv.Batch(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestServer_batch_invalidJSON(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/Batch", strings.NewReader(`{invalid-json}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 42)

	err := srv.Batch(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "error")
}

func TestServer_batch_empty(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/Batch", strings.NewReader(`[]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 42)

	err := srv.Batch(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "empty")
}

func TestServer_batch_saveError(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "https://a.com").
			Return(storage.Data{}, sql.ErrNoRows)

		mock.EXPECT().
			Save(gomock.Any(), gomock.AssignableToTypeOf(storage.Data{})).
			Return(errors.New("save error"))
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/Batch", strings.NewReader(
		`[{"correlation_id": "1", "original_url": "https://a.com"}]`,
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 42)

	err := srv.Batch(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "save error")
}

func TestServer_userURLs_success(t *testing.T) {
	userID := 42
	urls := []storage.Data{
		{OriginalURL: "https://a.com", ShortURL: "abc1"},
		{OriginalURL: "https://b.com", ShortURL: "abc2"},
	}

	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			GetAll(gomock.Any(), userID).
			Return(urls, nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", userID)

	err := srv.UserURLs(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "abc1")
	require.Contains(t, rec.Body.String(), "abc2")
	require.Contains(t, rec.Body.String(), srv.cfg.Server.FlagSuffixAddr)
}

func TestServer_userURLs_unauthorized(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", "not-int")

	err := srv.UserURLs(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Contains(t, rec.Body.String(), "userID is not of type int")
}

func TestServer_userURLs_noContent(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			GetAll(gomock.Any(), 42).
			Return(nil, ErrNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 42)

	err := srv.UserURLs(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestServer_userURLs_internalError(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			GetAll(gomock.Any(), 42).
			Return(nil, errors.New("db down"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 42)

	err := srv.UserURLs(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "getting data")
}

func TestServer_deleteUserURLs_bindError(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`{bad json}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 42)

	err := srv.DeleteUserURLs(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestServer_deleteUserURLs_emptyList(t *testing.T) {
	srv := setupTestServer(t, func(mock *mocks.MockDataKeeper) {})

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`[]`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := srv.echo.NewContext(req, rec)
	ctx.Set("userID", 42)

	err := srv.DeleteUserURLs(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, rec.Code)
}
