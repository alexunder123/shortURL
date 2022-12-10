package main

import (
	"log"
	"net/http"
	"shortURL/internal/handlers"
)

func main() {
	r := handlers.NewRouter()
	
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
