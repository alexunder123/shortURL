package storage

import (
	"encoding/json"
	"log"
	"os"
	"shortURL/internal/config"
	"sync"
)

type FileStorage struct {
	StorageStruct
}

func NewFileStorager(P *config.Param) Storager {
	ReadStorage(P)
	return &FileStorage{
		StorageStruct: StorageStruct{
			Key:   "",
			Value: "",
		},
	}
}

func (s *FileStorage) SetShortURL(fURL string, Params *config.Param) string {
	s.Key = HashStr(fURL)
	_, true := BaseURL[s.Key]
	if true {
		return s.Key
	}

	var mutex sync.Mutex
	mutex.Lock()
	BaseURL[s.Key] = fURL
	file, err := NewWriterFile(Params)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteFile(s.Key, fURL)
	mutex.Unlock()
	return s.Key
}

func (s *FileStorage) RetFullURL(key string) string {
	return BaseURL[key]
}

type readerFile struct {
	file    *os.File
	decoder *json.Decoder
}


func ReadStorage(P *config.Param) {
	file, err := NewReaderFile(P)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.ReadFile()
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

func (r *readerFile) ReadFile() {
	var fileBZ = make([]byte, 0)
	_, err := r.file.Read(fileBZ)
	if err != nil {
		log.Println(err)
		return
	}
	for r.decoder.More() {
		var t StorageStruct
		err := r.decoder.Decode(&t)
		if err != nil {
			log.Println(err)
			return
		}
		BaseURL[t.Key] = t.Value
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

func (w *writerFile) WriteFile(key, value string) {
	t := StorageStruct{Key: key, Value: value}
	err := w.encoder.Encode(&t)
	if err != nil {
		log.Println(err)
	}
}

func (w *writerFile) Close() error {
	return w.file.Close()
}
