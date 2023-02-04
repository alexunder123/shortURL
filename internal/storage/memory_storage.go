package storage

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
)

type MemoryStorage struct {
	baseURL    map[string]string
	userURL    map[string]string
	deletedURL map[string]bool
	sync.RWMutex
}

func NewMemoryStorager() Storager {
	return &MemoryStorage{
		baseURL:    make(map[string]string),
		userURL:    make(map[string]string),
		deletedURL: make(map[string]bool),
	}
}

func (s *MemoryStorage) SetShortURL(fURL, userID string, Params *config.Param) (string, error) {
	key := hashStr(fURL)
	s.RLock()
	id := s.userURL[key]
	s.RUnlock()
	if id == userID {
		return "", ErrConflict
	}
	s.Lock()
	s.baseURL[key] = fURL
	s.userURL[key] = userID
	s.deletedURL[key] = false
	s.Unlock()
	return key, nil
}

func (s *MemoryStorage) RetFullURL(key string) (string, error) {
	s.RLock()
	del := s.deletedURL[key]
	s.RUnlock()
	if del {
		return "", ErrGone
	}
	s.RLock()
	fURL, ok := s.baseURL[key]
	s.RUnlock()
	if !ok {
		return key, ErrNoContent
	}
	return fURL, nil
}

func (s *MemoryStorage) ReturnAllURLs(userID string, P *config.Param) ([]byte, error) {
	if len(s.baseURL) == 0 {
		return nil, ErrNoContent
	}
	var allURLs = make([]urls, 0)
	for key, value := range s.baseURL {
		s.RLock()
		id := s.userURL[key]
		s.RUnlock()
		if id == userID {
			allURLs = append(allURLs, urls{P.URL + "/" + key, value})
		}
	}
	if len(allURLs) == 0 {
		return nil, ErrNoContent
	}
	sb, err := json.Marshal(allURLs)
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
		key := hashStr(v.OriginURL)
		s.Lock()
		s.baseURL[key] = v.OriginURL
		s.userURL[key] = userID
		s.deletedURL[key] = false
		s.Unlock()
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(P.URL + "/" + key)
	}
	return r, nil
}

func (s *MemoryStorage) CloseDB() {
	log.Info().Msg("closed")
}

func (s *MemoryStorage) MarkDeleted(keys []string, id string) {
	s.Lock()
	for _, key := range keys {
		if s.userURL[key] == id {
			s.deletedURL[key] = true
		}
	}
	s.Unlock()
}
