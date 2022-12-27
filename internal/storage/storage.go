package storage

import (
	"crypto/md5"
	"fmt"
	"shortURL/internal/config"
)

var (
	BaseURL = make(map[string]string)
)

type Storager interface {
	SetShortURL(fURL string, Params *config.Param) string

	RetFullURL(key string) string
}

func NewStorager(P *config.Param) Storager {
	switch P.SaveFile {
	case 1:
		return NewFileStorager(P)
	default:
		return NewMemoryStorager()
	}
}

type StorageStruct struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func HashStr(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
