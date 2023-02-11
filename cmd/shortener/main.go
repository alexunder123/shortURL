package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
	"shortURL/internal/handler"
	"shortURL/internal/logger"
	"shortURL/internal/router"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

func main() {
	logger.Newlogger()
	log.Info().Msg("Start program")
	cnfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("NewConfig read environment error")
	}
	strg := storage.NewStorage(cnfg)
	log.Debug().Msg("storage init")
	deletingWorker := worker.NewWorker()
	hndlr := handler.NewHandler(cnfg, strg, deletingWorker)
	router := router.NewRouter(hndlr)
	log.Debug().Msg("handler init")
	deletingWorker.Run(strg, cnfg.DeletingBufferSize, cnfg.DeletingBufferTimeout)
	go func() {
		err := http.ListenAndServe(cnfg.ServerAddress, router)
		if err != nil {
			log.Fatal().Msgf("server failed: %s", err)
		}
	}()
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case sig := <-sigChan:
			log.Info().Msgf("OS cmd received signal %s", sig.String())
			deletingWorker.Stop()
			strg.CloseDB()
			os.Exit(0)

		}
	}

}
