package main

import (
	"log"
	"net/http"
	"shortURL/internal/config"
	"shortURL/internal/handlers"
	"shortURL/internal/storage"
)

func main() {
	params := config.NewEnv()
	storage := storage.NewStorager(params)
	r := handlers.NewRouter(params, storage)

	log.Fatal(http.ListenAndServe(params.Server, r))
}
