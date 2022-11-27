package main

import (
	"net/http"
	"shortURL/internal/handlers"
)

func main() {

	http.HandleFunc("/", handlers.ShortenerURL)
	http.ListenAndServe("localhost:8080", nil)
}
