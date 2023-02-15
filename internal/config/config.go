package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
)

type SaveMethod int

const (
	SaveMemory SaveMethod = iota
	SaveFile
	SaveSQL
)

type Config struct {
	ServerAddress         string `env:"SERVER_ADDRESS"`
	BaseURL               string `env:"BASE_URL"`
	FileStoragePath       string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN           string `env:"DATABASE_DSN"`
	SavePlace             SaveMethod
	DeletingBufferSize    int
	DeletingBufferTimeout time.Duration
}

func NewConfig() (*Config, error) {
	var config Config

	err := env.Parse(&config)
	if err != nil {
		return nil, err
	}

	if config.ServerAddress == "" {
		flag.StringVar(&config.ServerAddress, "a", "127.0.0.1:8080", "Адрес запускаемого сервера")
	}
	if config.BaseURL == "" {
		flag.StringVar(&config.BaseURL, "b", "http://127.0.0.1:8080", "Базовый адрес результирующего URL")
	}
	if config.FileStoragePath == "" {
		flag.StringVar(&config.FileStoragePath, "f", "", "Файловое хранилище URL")
	}
	if config.DatabaseDSN == "" {
		flag.StringVar(&config.DatabaseDSN, "d", "", "База данных SQL")
	}
	flag.Parse()

	if config.DatabaseDSN != "" {
		config.SavePlace = SaveSQL
	} else if config.FileStoragePath != "" {
		config.SavePlace = SaveFile
	}

	config.DeletingBufferSize = 10
	config.DeletingBufferTimeout = 100 * time.Millisecond

	return &config, nil
}
