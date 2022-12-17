package app

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Param struct {
	Server  string `env:"SERVER_ADDRESS"`
	URL     string `env:"BASE_URL"`
	Storage string `env:"FILE_STORAGE_PATH"`
	SaveDB  bool
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

func (P *Param) OpenDB() {
	if P.Storage == "" {
		return
	}
	file, err := NewReaderDB(P)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	file.ReadDB()
	P.SaveDB = true
}
