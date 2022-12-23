package storage

import (
	"shortURL/internal/config"
	"sync"
)

type MemoryStorage struct {
	StorageStruct
}

func NewMemoryStorager() Storager {
	return &MemoryStorage{
		StorageStruct: StorageStruct{
			Key:   "",
			Value: "",
		},
	}
}

func (s *MemoryStorage) SetShortURL(fURL string, Params *config.Param) string {
	s.Key = HashStr(fURL)
	_, true := BaseURL[s.Key]
	if true {
		return s.Key
	}

	var mutex sync.Mutex
	mutex.Lock()
	BaseURL[s.Key] = fURL
	mutex.Unlock()
	return s.Key
}

func (s *MemoryStorage) RetFullURL(key string) string {
	return BaseURL[key]
}
