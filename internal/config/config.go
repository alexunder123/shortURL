package config

import (
	"github.com/rs/zerolog/log"

	"github.com/caarlos0/env/v6"
)

type SaveMethod int

const (
	SaveMemory SaveMethod = iota
	SaveFile
	SaveSQL
)

type Param struct {
	Server    string `env:"SERVER_ADDRESS"`
	URL       string `env:"BASE_URL"`
	Storage   string `env:"FILE_STORAGE_PATH"`
	SQL       string `env:"DATABASE_DSN"`
	SavePlace SaveMethod
}

func NewConfig() *Param {
	var params Param

	err := env.Parse(&params)
	if err != nil {
		log.Fatal().Err(err).Msg("NewConfig read envinronment error")
	}
	ReadFlags(&params)
	if params.SQL != "" {
		params.SavePlace = SaveSQL
	} else if params.Storage != "" {
		params.SavePlace = SaveFile
	}

	return &params
}
