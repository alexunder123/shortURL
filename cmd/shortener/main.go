package main

import (
	"fmt"
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

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

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

	//требование statictest: "the channel used with signal.Notify should be buffered"
	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
loop:
	//требование statictest: "should use for range instead of for { select {} }"
	for sig := range sigChan {
		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			log.Info().Msgf("OS cmd received signal %s", sig)
			deletingWorker.Stop()
			strg.CloseDB()
			break loop
		}
	}

}
