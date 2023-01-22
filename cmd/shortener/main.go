package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortURL/internal/config"
	"shortURL/internal/handlers"
	"shortURL/internal/storage"
	"syscall"
)

func main() {
	log.Println("Start program")
	params := config.NewConfig()
	store := storage.NewStorage(params)
	log.Println("storage init")
	r := handlers.NewRouter(params, store)
	log.Println("handler init")
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range sigChan {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Println("Начинаем выход из программы")
				store.CloseDB()
				os.Exit(0)
			}
		}
	}()
	log.Fatal(http.ListenAndServe(params.Server, r.Router))
}
