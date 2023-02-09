package storage

import (
	"crypto/md5"
	"fmt"

	"shortURL/internal/config"
)

type Storager interface {
	SetShortURL(fURL, UserID string, Params *config.Param) (string, error)
	WriteMultiURL(m []MultiURL, UserID string, P *config.Param) ([]MultiURL, error)
	RetFullURL(key string) (string, error)
	ReturnAllURLs(UserID string, P *config.Param) ([]byte, error)
	CheckPing(P *config.Param) error
	CloseDB()
	MarkDeleted([]string, string)
}

// Тип возвращаемого значения - интерфейс
func NewStorage(P *config.Param) Storager {
	switch P.SavePlace {
	case config.SaveFile:
		return NewFileStorager(P)
	case config.SaveSQL:
		return NewSQLStorager(P)
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

type MultiURL struct {
	CorrID    string `json:"correlation_id"`
	OriginURL string `json:"original_url,omitempty"`
	ShortURL  string `json:"short_url,omitempty"`
}
