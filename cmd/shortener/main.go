package main

import (
	"log"
	"net/http"
	"shortURL/internal/config"
	"shortURL/internal/handlers"
	"shortURL/internal/storage"
	"time"
)

func main() {
	t1 := time.Now()
	params := config.NewConfig()
	log.Println("Config initialized")
	storage := storage.NewStorage(params)
	log.Println("Storage initialized")
	r := handlers.NewRouter(params, storage)
	log.Println("Server initialized")
	duration := time.Since(t1)
	log.Printf("Время инициализации %d мс \n", duration.Milliseconds())
	log.Fatal(http.ListenAndServe(params.Server, r))
}
