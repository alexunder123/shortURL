package main

import (
	"log"
	"net/http"
	"shortURL/internal/app"
	"shortURL/internal/handlers"
)

func main() {
	Params := app.GetEnv()
	r := handlers.NewRouter(Params)

	log.Fatal(http.ListenAndServe(Params.Server, r))
}
