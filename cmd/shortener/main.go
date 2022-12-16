package main

import (
	"log"
	"net/http"
	"shortURL/internal/app"
	"shortURL/internal/handlers"
)

func main() {
	Params := app.GetEnv()
	Params.OpenDB()
	r := handlers.NewRouter(Params)

	log.Fatal(http.ListenAndServe(Params.Server, r))
}
