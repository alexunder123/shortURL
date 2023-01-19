package storage

import (
	"crypto/md5"
	"errors"
	"fmt"
	"shortURL/internal/config"
)

var (
	BaseURL            = make(map[string]string)
	UserURL            = make(map[string]string)
	ErrNoContent error = errors.New("StatusNoContent")
	ErrConflict  error = errors.New("StatusConflict")
)

type Storager interface {
	SetShortURL(fURL, UserID string, Params *config.Param) (string, error)
	WriteMultiURL(m *[]MultiURL, UserID string, P *config.Param) (*[]MultiURL, error)
	RetFullURL(key string) string
	ReturnAllURLs(UserID string, P *config.Param) ([]byte, error)
	CheckPing(P *config.Param) error
	CloseDB()
}

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

type StorageStruct struct {
	UserID string `json:"ID"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func HashStr(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

type URLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type MultiURL struct {
	CorrID    string `json:"correlation_id"`
	OriginURL string `json:"original_url,omitempty"`
	ShortURL  string `json:"short_url,omitempty"`
}
