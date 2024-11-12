package main

import (
	"os"
	"syscall"

	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/server"
	"github.com/wickedv43/go-shortener/internal/storage"

	"github.com/samber/do/v2"
)

func main() {
	// provide part
	i := do.New()

	do.Provide(i, server.NewServer)
	do.Provide(i, config.NewConfig)
	do.Provide(i, storage.NewStorage)
	do.Provide(i, logger.NewLogger)

	do.MustInvoke[*logger.Logger](i)
	do.MustInvoke[*server.Server](i).Start()

	i.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
}
