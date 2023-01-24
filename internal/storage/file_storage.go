package storage

import (
	"encoding/json"
	"errors"
	"os"
	"shortURL/internal/config"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type FileStorage struct {
	baseURL map[string]string
	userURL map[string]string
	sync.RWMutex
}

func NewFileStorager(P *config.Param) Storager {
	fs := FileStorage{
		baseURL: make(map[string]string),
		userURL: make(map[string]string),
	}
	ReadStorage(P, &fs)
	return &fs
}

func (s *FileStorage) SetShortURL(fURL, UserID string, Params *config.Param) (string, error) {
	key := HashStr(fURL)
	_, true := s.baseURL[key]
	if true {
		if s.userURL[key] == UserID {
			return key, ErrConflict
		}
	}

	var mutex sync.Mutex
	mutex.Lock()
	BaseURL[key] = fURL
	UserURL[key] = UserID
	DeletedURL[key] = false
	mutex.Unlock()
	file, err := NewWriterFile(Params)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer file.Close()
	file.WriteFile(key, UserID, fURL)
	return key, nil
}

func (s *FileStorage) RetFullURL(key string) string {
	s.RLock()
	del := DeletedURL[key]
	s.RUnlock()
	if del {
		return "", ErrGone
	}
	
	return s.baseURL[key], nil
}

type readerFile struct {
	file    *os.File
	decoder *json.Decoder
}

func (s *FileStorage) ReturnAllURLs(UserID string, P *config.Param) ([]byte, error) {
	if len(s.baseURL) == 0 {
		return nil, ErrNoContent
	}
	var AllURLs = make([]URLs, 0)
	s.Lock()
	for key, value := range s.baseURL {
		if s.userURL[key] == UserID {
			AllURLs = append(AllURLs, URLs{P.URL + "/" + key, value})
		}
	}
	s.Unlock()
	if len(AllURLs) == 0 {
		return nil, ErrNoContent
	}
	sb, err := json.Marshal(AllURLs)
	if err != nil {
		return nil, err
	}
	return sb, nil
}

func (s *FileStorage) CheckPing(P *config.Param) error {
	return errors.New("wrong DB used: file storage")
}

func (s *FileStorage) WriteMultiURL(m []MultiURL, UserID string, P *config.Param) ([]MultiURL, error) {
	r := make([]MultiURL, len(m))
	file, err := NewWriterFile(P)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer file.Close()
	for i, v := range m {
		key := HashStr(v.OriginURL)
		var mutex sync.RWMutex
		mutex.Lock()
		BaseURL[Key] = v.OriginURL
		UserURL[Key] = UserID
		DeletedURL[s.Key] = false
		mutex.Unlock()
		file.WriteFile(key, UserID, v.OriginURL)
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(P.URL + "/" + key)
	}

	return r, nil
}

func ReadStorage(P *config.Param, fs *FileStorage) {
	file, err := NewReaderFile(P)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer file.Close()
	file.ReadFile(fs)
}

func NewReaderFile(P *config.Param) (*readerFile, error) {
	file, err := os.OpenFile(P.Storage, os.O_RDONLY|os.O_CREATE, 0777)
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
		log.Error().Err(err)
		return
	}
	for r.decoder.More() {
		var t StorageStruct
		err := r.decoder.Decode(&t)
		if err != nil {
			log.Error().Err(err)
			return
		}
		BaseURL[t.Key] = t.Value
		UserURL[t.Key] = t.UserID
		DeletedURL[t.Key] = t.Deleted
	}
}

func (r *readerFile) Close() error {
	return r.file.Close()
}

type writerFile struct {
	file    *os.File
	encoder *json.Encoder
}

func NewWriterFile(P *config.Param) (*writerFile, error) {
	file, err := os.OpenFile(P.Storage, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &writerFile{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (w *writerFile) WriteFile(key, userID, value string) {
	t := StorageStruct{UserID: userID, Key: key, Value: value, Deleted: false}
	err := w.encoder.Encode(&t)
	if err != nil {
		log.Error().Err(err)
	}
}

func (w *writerFile) Close() error {
	return w.file.Close()
}

func (s *FileStorage) CloseDB() {
	log.Info().Msg("file closed")
}
