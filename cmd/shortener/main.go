package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
	"shortURL/internal/handlers"
	"shortURL/internal/router"
	"shortURL/internal/storage"
)

func main() {
	setup()
	log.Info().Msg("Start program")
	params := config.NewConfig()
	store := storage.NewStorage(params)
	log.Debug().Msg("storage init")
	r := router.NewRouter(params, store)
	h := handlers.NewHandler(r)
	log.Debug().Msg("handler init")
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		ctxStop := r.ProcessingDel(ctx)
	loop:
		for s := range sigChan {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Info().Msg("Start exiting program")
				break loop
			}
		}
		cancel()
		<-ctxStop.Done()
		store.CloseDB()
		os.Exit(0)

	}()
	log.Fatal().Msgf("server failed: %s", http.ListenAndServe(params.Server, h))
}

func setup() {

	zerolog.TimeFieldFormat = ""

	zerolog.TimestampFunc = func() time.Time {
		return time.Date(2008, 1, 8, 17, 5, 05, 0, time.UTC)
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}
