package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
)

type FileStorage struct {
	baseURL    map[string]string
	userURL    map[string]string
	deletedURL map[string]bool
	sync.RWMutex
}

func NewFileStorager(cfg *config.Config) *FileStorage {
	fs := FileStorage{
		baseURL:    make(map[string]string),
		userURL:    make(map[string]string),
		deletedURL: make(map[string]bool),
	}
	ReadStorage(cfg, &fs)
	return &fs
}

func (s *FileStorage) SetShortURL(fURL, userID string, cfg *config.Config) (string, error) {
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
	file, err := NewWriterFile(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("SetShortURL NewWriterFile err")
	}
	defer file.Close()
	err = file.WriteFile(key, userID, fURL)
	return key, err
}

func (s *FileStorage) RetFullURL(key string) (string, error) {
	s.RLock()
	del := s.deletedURL[key]
	s.RUnlock()
	if del {
		return key, ErrGone
	}
	s.RLock()
	fURL, ok := s.baseURL[key]
	s.RUnlock()
	if !ok {
		return "", ErrNoContent
	}
	return fURL, nil
}

type readerFile struct {
	file    *os.File
	decoder *json.Decoder
}

func (s *FileStorage) ReturnAllURLs(userID string, cfg *config.Config) ([]byte, error) {
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
	sb, err := json.Marshal(allURLs)
	if err != nil {
		return nil, err
	}
	return sb, nil
}

func (s *FileStorage) CheckPing(P *config.Config) error {
	return errors.New("wrong DB used: file storage")
}

func (s *FileStorage) WriteMultiURL(m []MultiURL, userID string, cfg *config.Config) ([]MultiURL, error) {
	r := make([]MultiURL, len(m))
	file, err := NewWriterFile(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("WriteMultiURL NewWriterFile err")
	}
	defer file.Close()
	for i, v := range m {
		key := hashStr(v.OriginURL)
		s.Lock()
		s.baseURL[key] = v.OriginURL
		s.userURL[key] = userID
		s.deletedURL[key] = false
		s.Unlock()
		err := file.WriteFile(key, userID, v.OriginURL)
		if err != nil {
			return nil, err
		}
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(cfg.BaseURL + "/" + key)
	}

	return r, nil
}

func ReadStorage(cfg *config.Config, fs *FileStorage) {
	file, err := NewReaderFile(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("ReadStorage NewWriterFile err")
	}
	defer file.Close()
	file.ReadFile(fs)
}

func NewReaderFile(cfg *config.Config) (*readerFile, error) {
	file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &readerFile{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (r *readerFile) ReadFile(fs *FileStorage) {
	var fileBZ = make([]byte, 0)
	_, err := r.file.Read(fileBZ)
	if err != nil {
		log.Error().Err(err).Msg("ReadFile reading file err")
		return
	}
	for r.decoder.More() {
		var t storageStruct
		err := r.decoder.Decode(&t)
		if err != nil {
			log.Error().Err(err).Msg("ReadFile decoder err")
			return
		}
		fs.Lock()
		fs.baseURL[t.Key] = t.Value
		fs.userURL[t.Key] = t.UserID
		fs.deletedURL[t.Key] = t.Deleted
		fs.Unlock()
	}
}

func (r *readerFile) Close() error {
	return r.file.Close()
}

type writerFile struct {
	file    *os.File
	encoder *json.Encoder
}

func NewWriterFile(cfg *config.Config) (*writerFile, error) {
	file, err := os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &writerFile{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (w *writerFile) WriteFile(key, userID, value string) error {
	t := storageStruct{UserID: userID, Key: key, Value: value, Deleted: false}
	return w.encoder.Encode(&t)
}

func (w *writerFile) Close() error {
	return w.file.Close()
}

func (s *FileStorage) CloseDB() {
	log.Info().Msg("file closed")
}

func (s *FileStorage) MarkDeleted(keys []string, ids []string) {
	s.Lock()
	for i, key := range keys {
		if s.userURL[key] == ids[i] {
			s.deletedURL[key] = true
		}
	}
	s.Unlock()
}
