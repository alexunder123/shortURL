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

	// os.Setenv("SERVER_ADDRESS", "127.0.0.1:8080")
	// os.Setenv("BASE_URL", "http://127.0.0.1:8080/")

	err := env.Parse(&Params)
	if err != nil {
		log.Fatal(err)
	}
	if Params.Server == "" {
		Params.Server = "127.0.0.1:8080"
	}
	if Params.URL == "" {
		Params.URL = "http://" + Params.Server + "/"
	}
	return &Params
}
