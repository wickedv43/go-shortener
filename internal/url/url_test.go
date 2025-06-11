package url

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/mocks"
	"github.com/wickedv43/go-shortener/internal/storage"
	"go.uber.org/mock/gomock"
)

func setupTestURLService(tb testing.TB, configureMock func(*mocks.MockDataKeeper)) *URLService {
	tb.Helper()

	ctrl := gomock.NewController(tb)

	container := do.New()

	do.Provide(container, func(i do.Injector) (*config.Config, error) {
		return &config.Config{Server: config.Server{FlagSuffixAddr: "http://localhost:8080"}}, nil
	})
	do.Provide(container, logger.NewLogger)

	mockDataKeeper := mocks.NewMockDataKeeper(ctrl)
	configureMock(mockDataKeeper)

	do.Provide(container, func(i do.Injector) (storage.DataKeeper, error) {
		return mockDataKeeper, nil
	})

	do.Provide(container, NewURLService)

	return do.MustInvoke[*URLService](container)
}
