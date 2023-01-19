package main

import (
	"log"
	"net/http"
	"os"
	"shortURL/internal/app"
	"shortURL/internal/config"
	"shortURL/internal/handlers"
	"shortURL/internal/storage"
)

func main() {
	log.Println("Start program")
	params := config.NewConfig()
	store := storage.NewStorage(params)
	log.Println("storage init")
	r := handlers.NewRouter(params, store)
	log.Println("handler init")
	app.CloserDB(store)
	go func() {
		<-app.Stop
		os.Exit(0)
	}()
	log.Fatal(http.ListenAndServe(params.Server, r))
}
