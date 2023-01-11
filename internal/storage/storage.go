package storage

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"shortURL/internal/config"
	"syscall"
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
}

func NewStorage(P *config.Param) Storager {
	switch P.SaveFile {
	case 1:
		return NewFileStorager(P)
	case 2:
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

func CloserDB(P *config.Param) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range sigChan {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Println("Начинаем выход из программы")
				if P.SaveFile == 2 {
					CloseDB()
				} else {
					os.Exit(0)
				}
			}
		}
	}()
}
