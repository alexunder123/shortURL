package main

import (
	"log"
	"net/http"
	"shortURL/internal/config"
	"shortURL/internal/handlers"
	"shortURL/internal/storage"
)

func main() {
	params := config.NewConfig()
	storage := storage.NewStorage(params)
	r := handlers.NewRouter(params, storage)

	log.Fatal(http.ListenAndServe(params.Server, r))
}
