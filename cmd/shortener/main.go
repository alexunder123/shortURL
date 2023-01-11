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
	store := storage.NewStorage(params)
	log.Println("storage init")
	r := handlers.NewRouter(params, store)
	log.Println("handler init")
	defer func() {
		log.Println("defer ")
		storage.CloseDB()
	} ()
	log.Fatal(http.ListenAndServe(params.Server, r))
}
