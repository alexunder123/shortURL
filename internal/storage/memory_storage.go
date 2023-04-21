package storage

import (
	"errors"
	"sync"

	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
)

// MemoryStorage структура для хранения данных в оперативной памяти.
type MemoryStorage struct {
	baseURL    map[string]string
	userURL    map[string]string
	deletedURL map[string]bool
	sync.RWMutex
}

// NewMemoryStorager метод генерирует хранилище данных.
func NewMemoryStorager() *MemoryStorage {
	return &MemoryStorage{
		baseURL:    make(map[string]string),
		userURL:    make(map[string]string),
		deletedURL: make(map[string]bool),
	}
}

// SetShortURL метод генерирует ключ для короткой ссылки, проверяет его наличие и сохраняет данные.
// Данные передаются и возвращаются текстом в теле запроса.
func (s *MemoryStorage) SetShortURL(fURL string, userID string, cfg *config.Config) (string, error) {
	key := hashStr(fURL)
	s.RLock()
	id := s.userURL[key]
	s.RUnlock()
	if id == userID {
		return cfg.BaseURL + "/" + key, ErrConflict
	}
	s.Lock()
	s.baseURL[key] = fURL
	s.userURL[key] = userID
	s.deletedURL[key] = false
	s.Unlock()
	return cfg.BaseURL + "/" + key, nil
}

// RetFullURL метод возвращает полный адрес по ключу от короткой ссылки.
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

// ReturnAllURLs метод возвращает список сокращенных адресов по ID пользователя.
func (s *MemoryStorage) ReturnAllURLs(userID string, cfg *config.Config) ([]urls, error) {
	if len(s.baseURL) == 0 {
		return nil, ErrNoContent
	}
	var allURLs = make([]urls, 0)
	for key, value := range s.baseURL {
		s.RLock()
		id := s.userURL[key]
		s.RUnlock()
		if id == userID {
			allURLs = append(allURLs, urls{cfg.BaseURL + "/" + key, value})
		}
	}
	if len(allURLs) == 0 {
		return nil, ErrNoContent
	}
	return allURLs, nil
}

// CheckPing метод возвращает статус подключения к базе данных.
func (s *MemoryStorage) CheckPing(cfg *config.Config) error {
	return errors.New("wrong DB used: memory storage")
}

// WriteMultiURL метод обрабатывает, сохраняет и возвращает batch список сокращенных адресов.
func (s *MemoryStorage) WriteMultiURL(m []MultiURL, userID string, cfg *config.Config) ([]MultiURL, error) {
	r := make([]MultiURL, len(m))
	for i, v := range m {
		key := hashStr(v.OriginURL)
		s.Lock()
		s.baseURL[key] = v.OriginURL
		s.userURL[key] = userID
		s.deletedURL[key] = false
		s.Unlock()
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(cfg.BaseURL + "/" + key)
	}
	return r, nil
}

// CloseDB метод закрывает соединение с хранилищем данных.
func (s *MemoryStorage) CloseDB() {
	log.Info().Msg("closed")
}

// MarkDeleted метод помечает на удаление адреса пользователя в хранилище.
func (s *MemoryStorage) MarkDeleted(keys []string, ids []string) {
	s.Lock()
	for i, key := range keys {
		if s.userURL[key] == ids[i] {
			s.deletedURL[key] = true
		}
	}
	s.Unlock()
}

// ReturnStats метод возвращает статистику по количеству сохраненных сокращенных URL и пользователей.
func (s *MemoryStorage) ReturnStats() (*stats, error) {
	temp := make(map[string]bool)
	for _, v := range s.userURL {
		if !temp[v] {
			temp[v] = true
		}
	}
	stats := stats{
		URLs:  len(s.baseURL),
		Users: len(temp),
	}
	return &stats, nil
}
