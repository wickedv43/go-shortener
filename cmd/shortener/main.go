package main

import (
	"os"
	"syscall"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/server"
	"github.com/wickedv43/go-shortener/internal/storage"
)

func main() {
	i := do.New()

	do.Provide(i, server.NewServer)
	do.Provide(i, config.NewConfig)
	do.Provide(i, logger.NewLogger)

	do.Provide(i, storage.NewFileStorage)
	do.Provide(i, storage.NewLocalStorage)
	do.Provide(i, storage.NewPostgresStorage)

	provideStorageByPriority(i)

	do.MustInvoke[*server.Server](i).Start()

	i.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
}

func provideStorageByPriority(i do.Injector) {
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "storage")

	// Пробуем Postgres
	if s, err := do.Invoke[*storage.PostgresStorage](i); err == nil {
		log.Info("using postgres storage")
		do.Provide(i, func(i do.Injector) (storage.DataKeeper, error) {
			return s, nil
		})
		return
	}

	// Пробуем файл
	if s, err := do.Invoke[*storage.FileStorage](i); err == nil {
		log.Info("using file storage")
		do.Provide(i, func(i do.Injector) (storage.DataKeeper, error) {
			return s, nil
		})
		return
	}

	// Fallback: память
	s := do.MustInvoke[*storage.LocalStorage](i)
	log.Info("using memory storage")
	do.Provide(i, func(i do.Injector) (storage.DataKeeper, error) {
		return s, nil
	})
}
