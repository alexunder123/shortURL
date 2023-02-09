package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type SaveMethod int

const (
	saveMemory SaveMethod = iota
	saveFile
	saveSQL
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	SavePlace       SaveMethod
}

func NewConfig() (*Config, error) {
	var config *Config

	err := env.Parse(config)
	if err != nil {
		return nil, err
		//log.Fatal().Err(err).Msg("NewConfig read environment error")
	}

	if config.ServerAddress == "" {
		flag.StringVar(&config.ServerAddress, "a", "127.0.0.1:8080", "Адрес запускаемого сервера")
	}
	if config.BaseURL == "" {
		flag.StringVar(&config.BaseURL, "b", "http://127.0.0.1:8080", "Базовый адрес результирующего URL")
	}
	if config.FileStoragePath == "" {
		flag.StringVar(&config.FileStoragePath, "f", "", "Хранилище URL")
	}
	if config.DatabaseDSN == "" {
		flag.StringVar(&config.DatabaseDSN, "d", "", "База данных SQL")
	}
	flag.Parse()

	//if config.DatabaseDSN != "" {
	//	config.SavePlace = saveSQL
	//}
	//
	//if config.FileStoragePath != "" {
	//	config.SavePlace = saveFile
	//}

	return config, nil
}
