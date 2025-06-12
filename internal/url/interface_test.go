package url

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/mocks"
	"github.com/wickedv43/go-shortener/internal/storage"
	"go.uber.org/mock/gomock"
)

func TestURLService_Save(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().Get(gomock.Any(), "http://example.com").Return(storage.Data{}, errors.New("not found"))
		mock.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
	})

	ctx := context.Background()
	userID := 123
	originalURL := "http://example.com"

	data, err := service.Save(ctx, originalURL, userID)

	require.NoError(t, err)
	require.Equal(t, originalURL, data.OriginalURL)
	require.NotEmpty(t, data.ShortURL)
	require.Equal(t, userID, data.UUID)
}

func TestURLService_Save_Conflict(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().Get(gomock.Any(), "http://example.com").Return(storage.Data{
			OriginalURL: "http://example.com",
			ShortURL:    "short123",
			UUID:        123,
		}, nil)
	})

	ctx := context.Background()
	userID := 123
	originalURL := "http://example.com"

	_, err := service.Save(ctx, originalURL, userID)

	require.ErrorIs(t, err, ErrConflict)
}

func TestURLService_Get(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().Get(gomock.Any(), "short123").Return(storage.Data{
			OriginalURL: "http://example.com",
			ShortURL:    "short123",
			UUID:        123,
		}, nil)
	})

	ctx := context.Background()
	data, err := service.Get(ctx, "short123")

	require.NoError(t, err)
	require.Equal(t, "short123", data.ShortURL)
	require.Equal(t, "http://example.com", data.OriginalURL)
}

func TestURLService_Get_NotFound(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().Get(gomock.Any(), "notfound").Return(storage.Data{}, errors.New("not found"))
	})

	ctx := context.Background()
	_, err := service.Get(ctx, "notfound")

	require.Error(t, err)
}

func TestURLService_GetAll(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().GetAll(gomock.Any(), 123).Return([]storage.Data{
			{OriginalURL: "http://example.com", ShortURL: "short123", UUID: 123},
			{OriginalURL: "http://example.org", ShortURL: "short456", UUID: 123},
		}, nil)
	})

	ctx := context.Background()
	data, err := service.GetAll(ctx, 123)

	require.NoError(t, err)
	require.Len(t, data, 2)
	require.Contains(t, data[0].ShortURL, "http://localhost:8080/")
	require.Contains(t, data[1].ShortURL, "http://localhost:8080/")
}

func TestURLService_GetAll_NoContent(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().GetAll(gomock.Any(), 123).Return([]storage.Data{}, nil)
	})

	ctx := context.Background()
	_, err := service.GetAll(ctx, 123)

	require.ErrorIs(t, err, ErrNoContent)
}

func TestURLService_DeleteBatch(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().BatchDelete(gomock.Any()).Return(nil)
	})

	err := service.BatchDelete([]string{"short123", "short456"})

	require.NoError(t, err)
}

func TestURLService_DeleteUserURLS(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().Get(gomock.Any(), "short123").Return(storage.Data{
			OriginalURL: "http://example.com",
			ShortURL:    "short123",
			UUID:        123,
			DeletedFlag: false,
		}, nil)
		mock.EXPECT().Get(gomock.Any(), "short456").Return(storage.Data{
			OriginalURL: "http://example.org",
			ShortURL:    "short456",
			UUID:        123,
			DeletedFlag: false,
		}, nil)
		mock.EXPECT().BatchDelete(gomock.Any()).AnyTimes().Return(nil)
	})

	ctx := context.Background()
	err := service.DeleteUserURLS(ctx, 123, []string{"short123", "short456"})

	require.NoError(t, err)
}

func TestURLService_Stats(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().Stats().Return(42, 5, nil)
	})

	urls, users, err := service.Stats()

	require.NoError(t, err)
	require.Equal(t, 42, urls)
	require.Equal(t, 5, users)
}

func TestURLService_HealthCheck(t *testing.T) {
	service := setupTestURLService(t, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().HealthCheck().Return(nil)
	})

	err := service.Ping()

	require.NoError(t, err)
}
