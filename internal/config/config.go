package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Param struct {
	Server   string `env:"SERVER_ADDRESS"`
	URL      string `env:"BASE_URL"`
	Storage  string `env:"FILE_STORAGE_PATH"`
	SQL      string `env:"DATABASE_DSN"`
	SaveFile int
}

func NewConfig() *Param {
	var Params Param

	err := env.Parse(&Params)
	if err != nil {
		log.Fatal(err)
	}
	ReadFlags(&Params)

	if Params.Storage != "" {
		Params.SaveFile = 1
	}
	//В случае, если переданы и файловое и SQL хранилища. Приоритет будет у SQL
	if Params.SQL != "" {
		Params.SaveFile = 2
	}

	return &Params
}
