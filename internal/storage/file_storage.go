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
	baseURL    map[string]string
	userURL    map[string]string
	deletedURL map[string]bool
	sync.RWMutex
}

func NewFileStorager(P *config.Param) Storager {
	fs := FileStorage{
		baseURL:    make(map[string]string),
		userURL:    make(map[string]string),
		deletedURL: make(map[string]bool),
	}
	ReadStorage(P, &fs)
	return &fs
}

func (s *FileStorage) SetShortURL(fURL, userID string, Params *config.Param) (string, error) {
	key := HashStr(fURL)
	_, true := s.baseURL[key]
	if true {
		if s.userURL[key] == userID {
			return key, ErrConflict
		}
	}

	s.RLock()
	s.baseURL[key] = fURL
	s.userURL[key] = userID
	s.deletedURL[key] = false
	s.RUnlock()
	file, err := NewWriterFile(Params)
	if err != nil {
		log.Fatal().Err(err).Msg("SetShortURL NewWriterFile err")
	}
	defer file.Close()
	file.WriteFile(key, userID, fURL)
	return key, nil
}

func (s *FileStorage) RetFullURL(key string) (string, error) {
	s.RLock()
	del := s.deletedURL[key]
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

func (s *FileStorage) ReturnAllURLs(userID string, P *config.Param) ([]byte, error) {
	if len(s.baseURL) == 0 {
		return nil, ErrNoContent
	}
	var allURLs = make([]URLs, 0)
	s.Lock()
	for key, value := range s.baseURL {
		if s.userURL[key] == userID {
			allURLs = append(allURLs, URLs{P.URL + "/" + key, value})
		}
	}
	s.Unlock()
	if len(allURLs) == 0 {
		return nil, ErrNoContent
	}
	sb, err := json.Marshal(allURLs)
	if err != nil {
		return nil, err
	}
	return sb, nil
}

func (s *FileStorage) CheckPing(P *config.Param) error {
	return errors.New("wrong DB used: file storage")
}

func (s *FileStorage) WriteMultiURL(m []MultiURL, userID string, P *config.Param) ([]MultiURL, error) {
	r := make([]MultiURL, len(m))
	file, err := NewWriterFile(P)
	if err != nil {
		log.Fatal().Err(err).Msg("WriteMultiURL NewWriterFile err")
	}
	defer file.Close()
	for i, v := range m {
		key := HashStr(v.OriginURL)
		s.Lock()
		s.baseURL[key] = v.OriginURL
		s.userURL[key] = userID
		s.deletedURL[key] = false
		s.Unlock()
		file.WriteFile(key, userID, v.OriginURL)
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(P.URL + "/" + key)
	}

	return r, nil
}

func ReadStorage(P *config.Param, fs *FileStorage) {
	file, err := NewReaderFile(P)
	if err != nil {
		log.Fatal().Err(err).Msg("ReadStorage NewWriterFile err")
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
		log.Error().Err(err).Msg("ReadFile reading file err")
		return
	}
	for r.decoder.More() {
		var t StorageStruct
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

func (s *FileStorage) MarkDeleted(deleteURLs []string, UserID string, P *config.Param) {
	inputCh := make(chan string)
	go func() {
		for _, del := range deleteURLs {
			inputCh <- del
		}
		close(inputCh)
	}()

	chQ := 5 //Количество каналов для работы
	//fan out
	fanOutChs := fanOut(inputCh, chQ)
	workerChs := make([]chan string, 0, chQ)
	for _, fanOutCh := range fanOutChs {
		workerCh := make(chan string)
		s.newWorker(fanOutCh, workerCh, UserID)
		workerChs = append(workerChs, workerCh)
	}

	//fan in
	outCh := fanIn(workerChs)

	//update
	for key := range outCh {
		s.deletedURL[key] = true
	}
}

func (s *FileStorage) newWorker(in, out chan string, UserID string) {
	go func() {
		for myURL := range in {
			key := strings.Trim(myURL, "\"")
			s.RLock()
			id := s.userURL[key]
			s.RUnlock()
			if id != UserID {
				continue
			}
			out <- key
		}
		close(out)
	}()
}
