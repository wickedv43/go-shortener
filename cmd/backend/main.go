package main

import (
	"github.com/samber/do/v2"
	"github.com/wickedv43/go-shortener/cmd/config"
	"github.com/wickedv43/go-shortener/cmd/shortener"
	"github.com/wickedv43/go-shortener/cmd/storage"
	"os"
	"syscall"
)

func main() {
	// provide part
	i := do.New()

	do.Provide(i, server.NewServer)
	do.Provide(i, config.NewConfig)
	do.Provide(i, storage.NewStorage)

	do.MustInvoke[*server.Server](i).Start()

	i.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
}
