package main

import (
	"log"
	"net/http"
	"shortURL/internal/handlers"
)

func main() {
	r := handlers.NewRouter()

	// http.HandleFunc("/", handlers.ShortenerURL)
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
