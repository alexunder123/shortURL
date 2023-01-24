package storage

import (
	"crypto/md5"
	"errors"
	"fmt"
	"shortURL/internal/config"
	"sync"
)

var (
	BaseURL            = make(map[string]string)
	UserURL            = make(map[string]string)
	DeletedURL         = make(map[string]bool)
	ErrNoContent error = errors.New("StatusNoContent")
	ErrConflict  error = errors.New("StatusConflict")
	ErrGone      error = errors.New("StatusGone")
)

type Storager interface {
	SetShortURL(fURL, UserID string, Params *config.Param) (string, error)
	WriteMultiURL(m *[]MultiURL, UserID string, P *config.Param) (*[]MultiURL, error)
	RetFullURL(key string) (string, error)
	ReturnAllURLs(UserID string, P *config.Param) ([]byte, error)
	CheckPing(P *config.Param) error
	CloseDB()
	MarkDeleted(DeleteURLs *[]string, UserID string, P *config.Param)
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
	Deleted bool `json:"deleted"`
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

func fanOut(inputCh chan string, chQ int) []chan string {
	chs := make([]chan string, 0, chQ)
	for i := 0; i < chQ; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan string) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			num, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- num
		}
	}()
	return chs
}

func fanIn(inputChs []chan string) chan string {
	outCh := make(chan string)
	go func() {
		wg := &sync.WaitGroup{}
		for _, inputCh := range inputChs {
			wg.Add(1)
			go func(inputCh chan string) {
				defer wg.Done()
				for item := range inputCh {
					outCh <- item
				}
			}(inputCh)
		}
		wg.Wait()
		close(outCh)
	}()
	return outCh
}
