package storage

import (
	"encoding/json"
	"shortURL/internal/config"
	"sync"
)

type MemoryStorage struct {
	StorageStruct
}

func NewMemoryStorager() Storager {
	return &MemoryStorage{
		StorageStruct: StorageStruct{
			UserID: "",
			Key:    "",
			Value:  "",
		},
	}
}

func (s *MemoryStorage) SetShortURL(fURL, UserID string, Params *config.Param) string {
	s.Key = HashStr(fURL)
	_, true := BaseURL[s.Key]
	if true {
		return s.Key
	}

	var mutex sync.RWMutex
	mutex.Lock()
	BaseURL[s.Key] = fURL
	UserURL[s.Key] = UserID
	mutex.Unlock()
	return s.Key
}

func (s *MemoryStorage) RetFullURL(key string) string {
	return BaseURL[key]
}

func (s *MemoryStorage) ReturnAllURLs(UserID string, P *config.Param) ([]byte, error) {
	if len(BaseURL) == 0 {
		return nil, ErrNoContent
	}
	var AllURLs = make([]URLs, 0)
	var mutex sync.Mutex
	mutex.Lock()
	for key, value := range BaseURL {
		if UserURL[key] == UserID {
			AllURLs = append(AllURLs, URLs{P.URL + "/" + key, value})
		}
	}
	mutex.Unlock()
	if len(AllURLs) == 0 {
		return nil, ErrNoContent
	}
	sb, err := json.Marshal(AllURLs)
	if err != nil {
		return nil, err
	}
	return sb, nil
}
