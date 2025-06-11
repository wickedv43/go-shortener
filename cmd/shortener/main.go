//go:generate go run generator/build.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/server"
	"github.com/wickedv43/go-shortener/internal/storage"
	"github.com/wickedv43/go-shortener/internal/url"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i := do.New()

	do.Provide(i, server.NewServer)
	do.Provide(i, config.NewConfig)
	do.Provide(i, logger.NewLogger)

	do.Provide(i, storage.NewFileStorage)
	do.Provide(i, storage.NewLocalStorage)
	do.Provide(i, storage.NewPostgresStorage)

	do.Provide(i, func(i do.Injector) (url.Shortener, error) {
		return url.NewURLService(i)
	})

	log := do.MustInvoke[*logger.Logger](i)
	info := fmt.Sprintf("\n-----------------------\n"+
		"Build Version: %s\n"+
		"Build Date: %s\n"+
		"Build Commit: %s\n"+
		"-----------------------\n", buildVersion, buildDate, buildCommit)
	log.Info(info)

	provideStorageByPriority(i)

	flag.Parse()

	go do.MustInvoke[*server.Server](i).Start()

	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt}

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
