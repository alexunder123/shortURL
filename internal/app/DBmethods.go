package app

import (
	"encoding/json"
	"log"
	"os"
)

var (
	localBase DB
)

type DB struct {
	BaseURL map[string]string
}

type DBstring struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type readerDB struct {
	file    *os.File
	decoder *json.Decoder
}

func NewReaderDB(P *Param) (*readerDB, error) {
	file, err := os.OpenFile(P.Storage, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &readerDB{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (r *readerDB) ReadDB() {
	var fileBZ = make([]byte, 0)
	_, err := r.file.Read(fileBZ)
	if err != nil {
		log.Println(err)
		return
	}
	for r.decoder.More() {
		var t DBstring
		err := r.decoder.Decode(&t)
		if err != nil {
			log.Println(err)
			return
		}
		localBase.BaseURL[t.Key] = t.Value
	}
}

func (r *readerDB) Close() error {
	return r.file.Close()
}

type writerDB struct {
	file    *os.File
	encoder *json.Encoder
}

func NewWriterDB(P *Param) (*writerDB, error) {
	file, err := os.OpenFile(P.Storage, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &writerDB{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (w *writerDB) WriteDB(key, value string) {
	t := DBstring{Key: key, Value: value}
	err := w.encoder.Encode(&t)
	if err != nil {
		log.Println(err)
	}
}

func (w *writerDB) Close() error {
	return w.file.Close()
}
