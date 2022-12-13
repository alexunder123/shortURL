package app

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Param struct {
	Server string `env:"SERVER_ADDRESS"`
	URL    string `env:"BASE_URL"`
}

func GetEnv() *Param {
	var Params Param

	err := env.Parse(&Params)
	if err != nil {
		log.Fatal(err)
	}
	if Params.Server == "" {
		Params.Server = "127.0.0.1:8080"
	}
	if Params.URL == "" {
		Params.URL = "http://" + Params.Server
	}
	return &Params
}
