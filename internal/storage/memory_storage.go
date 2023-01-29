package storage

import (
	"encoding/json"
	"errors"
	"shortURL/internal/config"
	"sync"

	"github.com/rs/zerolog/log"
)

type MemoryStorage struct {
	baseURL map[string]string
	userURL map[string]string
	sync.RWMutex
}

func NewMemoryStorager() Storager {
	return &MemoryStorage{
		baseURL: make(map[string]string),
		userURL: make(map[string]string),
	}
}

func (s *MemoryStorage) SetShortURL(fURL, userID string, Params *config.Param) (string, error) {
	key := HashStr(fURL)
	_, true := s.baseURL[key]
	if true {
		if s.userURL[key] == userID {
			return key, ErrConflict
		}
	}

	s.Lock()
	s.baseURL[key] = fURL
	s.userURL[key] = userID
	s.Unlock()
	return key, nil
}

func (s *MemoryStorage) RetFullURL(key string) (string, error) {
	return s.baseURL[key], nil
}

func (s *MemoryStorage) ReturnAllURLs(userID string, P *config.Param) ([]byte, error) {
	if len(s.baseURL) == 0 {
		return nil, ErrNoContent
	}
	var AllURLs = make([]URLs, 0)
	var mutex sync.Mutex
	mutex.Lock()
	for key, value := range s.baseURL {
		if s.userURL[key] == userID {
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

func (s *MemoryStorage) CheckPing(P *config.Param) error {
	return errors.New("wrong DB used: memory storage")
}

func (s *MemoryStorage) WriteMultiURL(m []MultiURL, userID string, P *config.Param) ([]MultiURL, error) {
	r := make([]MultiURL, len(m))
	for i, v := range m {
		Key := HashStr(v.OriginURL)
		s.Lock()
		s.baseURL[Key] = v.OriginURL
		s.userURL[Key] = userID
		s.Unlock()
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(P.URL + "/" + Key)
	}

	return r, nil
}

func (s *MemoryStorage) CloseDB() {
	log.Info().Msg("closed")
}
