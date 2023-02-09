package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
	"shortURL/internal/handlers"
	"shortURL/internal/router"
	"shortURL/internal/storage"
)

type MyServer struct {
	http.Server
}

func main() {
	// Создать пакет logger, с конструктором NewLogger()
	// setup()

	ctx, cancel := context.WithCancel(context.Background())

	log.Info().Msg("Start program")

	config, err := config.NewConfig()
	if err != nil {
		// fatal
	}

	store := storage.NewStorage(cfg)
	//store, err := storage.NewStorage(cfg)
	//if err != nil {
	// log.Fatal(err)
	//}

	log.Debug().Msg("storage init")

	routing := router.NewRouter(config, store)

	log.Debug().Msg("handler init")
	handler := handlers.NewHandler(routing)

	// Новый пакет для удаления
	// deletingWorker, err := NewDeletingWorker()
	// if err != nil {
	//    log.Fatal(err)
	// }

	// !!!!!!! Обязательно воркер удаления
	go deletingWorker.Run(ctx)

	// Запуск сервера
	go func() {
		err = http.ListenAndServe(config.ServerAddress, handler)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}()

	// Завершение программы
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	for {
		select {
		case sig := <-signals:
			cancel()
			log.Printf("OS cmd received signal %s", sig.String())

			//
			deletingWorker.Stop()

			store.CloseDB()

			log.Info().Msg("application shutdown gracefully")
			os.Exit(0)
	}
}

// В пакет logger, setup в конструктор NewLogger()
//func setup() {
//
//	zerolog.TimeFieldFormat = ""
//
//	zerolog.TimestampFunc = func() time.Time {
//		return time.Date(2008, 1, 8, 17, 5, 05, 0, time.UTC)
//	}
//	zerolog.SetGlobalLevel(zerolog.InfoLevel)
//	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
//}
