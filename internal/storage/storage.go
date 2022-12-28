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
)

type Storager interface {
	SetShortURL(fURL, UserID string, Params *config.Param) string

	RetFullURL(key string) string
	ReturnAllURLs(UserID string, P *config.Param) ([]byte, error)
}

func NewStorage(P *config.Param) Storager {
	switch P.SaveFile {
	case 1:
		return NewFileStorager(P)
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
