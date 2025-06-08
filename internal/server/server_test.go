package server

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/mocks"
	"github.com/wickedv43/go-shortener/internal/url"
	"go.uber.org/mock/gomock"
)

func setupTestServer(tb testing.TB, configureMock func(*mocks.MockShortener)) *Server {
	tb.Helper()

	ctrl := gomock.NewController(tb)

	container := do.New()

	do.Provide(container, func(i do.Injector) (*config.Config, error) {
		return &config.Config{Server: config.Server{FlagRunAddr: ":8080"}}, nil
	})
	do.Provide(container, logger.NewLogger)

	mockShortener := mocks.NewMockShortener(ctrl)
	configureMock(mockShortener)

	do.Provide(container, func(i do.Injector) (url.Shortener, error) {
		return mockShortener, nil
	})

	do.Provide(container, NewServer)

	return do.MustInvoke[*Server](container)
}
