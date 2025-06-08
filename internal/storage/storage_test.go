package storage

import (
	"context"
	"os"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
)

func runDataKeeperComplianceTests(t *testing.T, name string, keeper DataKeeper) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		defer keeper.Close()

		ctx := context.Background()
		data := Data{
			UUID:        123,
			ShortURL:    "test123",
			OriginalURL: "http://example.com",
		}

		require.NoError(t, keeper.Save(ctx, data), "Save should not fail")

		got, err := keeper.Get(ctx, "test123")
		require.NoError(t, err)
		require.Equal(t, data.OriginalURL, got.OriginalURL)

		got2, err := keeper.Get(ctx, data.OriginalURL)
		require.NoError(t, err)
		require.Equal(t, data.ShortURL, got2.ShortURL)

		all, err := keeper.GetAll(ctx, 123)
		require.NoError(t, err)
		require.Len(t, all, 1)

		require.NoError(t, keeper.BatchDelete([]string{"test123"}))

		require.NoError(t, keeper.HealthCheck())
	})
}

func TestDataKeeperImplementations(t *testing.T) {
	t.Run("LocalStorage", func(t *testing.T) {
		st := &LocalStorage{}
		runDataKeeperComplianceTests(t, "LocalStorage", st)
	})

	t.Run("FileStorage", func(t *testing.T) {

		st := newTestFileStorage(t)

		runDataKeeperComplianceTests(t, "FileStorage", st)

	})

	t.Run("PostgresStorage", func(t *testing.T) {
		dsn := os.Getenv("TEST_DATABASE_DSN")
		if dsn == "" {
			t.Skip("TEST_DATABASE_DSN not set")
		}

		i := do.New()
		do.Provide(i, func(i do.Injector) (*config.Config, error) {
			return &config.Config{
				Server: config.Server{FlagDatabaseDSN: dsn},
			}, nil
		})
		do.Provide(i, logger.NewLogger)

		st, err := NewPostgresStorage(i)
		require.NoError(t, err)

		runDataKeeperComplianceTests(t, "PostgresStorage", st)
	})
}
