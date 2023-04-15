package storage

import (
	"crypto/md5"
	"fmt"

	"shortURL/internal/config"
)

// Storager - интерфейс для работы с хранилищем.
type Storager interface {
	SetShortURL(fURL string, userID string, cfg *config.Config) (string, error)
	SetShortURLjs(bytes []byte, userID string, cfg *config.Config) ([]byte, error)
	WriteMultiURL(bytes []byte, UserID string, P *config.Config) ([]byte, error)
	RetFullURL(key string) (string, error)
	ReturnAllURLs(UserID string, P *config.Config) ([]byte, error)
	ReturnStats() ([]byte, error)
	CheckPing(P *config.Config) error
	CloseDB()
	MarkDeleted([]string, []string)
}

// NewStorage метод "Фабрика" для создания хранилища в соответствии с конфигурацией сервиса.
func NewStorage(cfg *config.Config) Storager {
	switch cfg.SavePlace {
	case config.SaveFile:
		return NewFileStorager(cfg)
	case config.SaveSQL:
		return NewSQLStorager(cfg)
	default:
		return NewMemoryStorager()
	}
}

type storageStruct struct {
	UserID  string `json:"ID"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Deleted bool   `json:"deleted"`
}

func hashStr(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

type urls struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// MultiURL структура для обработки batch запросов в формате JSON.
type MultiURL struct {
	CorrID    string `json:"correlation_id"`
	OriginURL string `json:"original_url,omitempty"`
	ShortURL  string `json:"short_url,omitempty"`
}

type stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

type postURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}
