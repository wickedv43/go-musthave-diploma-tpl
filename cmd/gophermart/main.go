package main

import (
	"flag"
	"os"
	"syscall"

	"github.com/samber/do/v2"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/config"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/logger"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/server"
	"github.com/wickedv43/go-musthave-diploma-tpl/internal/storage"
)

func main() {
	flag.Parse()
	// provide part
	i := do.New()

	do.Provide(i, server.NewServer)
	do.Provide(i, config.NewConfig)
	do.Provide(i, logger.NewLogger)

	//storage
	do.Provide[storage.DataKeeper](i, func(i do.Injector) (storage.DataKeeper, error) {
		return do.MustInvoke[*storage.PostgresStorage](i), nil
	})

	do.MustInvoke[*logger.Logger](i)
	do.MustInvoke[*server.Server](i).Start()

	i.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
}
