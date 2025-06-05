//go:generate go run generator/build.go
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/server"
	"github.com/wickedv43/go-shortener/internal/storage"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	i := do.New()

	do.Provide(i, server.NewServer)
	do.Provide(i, config.NewConfig)
	do.Provide(i, logger.NewLogger)

	do.Provide(i, storage.NewFileStorage)
	do.Provide(i, storage.NewLocalStorage)
	do.Provide(i, storage.NewPostgresStorage)

	log := do.MustInvoke[*logger.Logger](i)
	log.Info("BuildVersion: ", buildVersion)
	log.Info("BuildDate: ", buildDate)
	log.Info("BuildCommit: ", buildCommit)

	provideStorageByPriority(i)

	flag.Parse()

	do.MustInvoke[*server.Server](i).Start()

	signalCh := make(chan os.Signal, 1)
	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt}

	signal.Notify(signalCh, signals...)
	<-signalCh

	cancel()

	_, err := i.ShutdownOnSignalsWithContext(ctx, signals...)
	if err != nil {
		log.Error("shutdown", err)
	}

	log.Info("grace shutdown")
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
