package storage

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"shortURL/internal/config"
	"strings"
	"sync"
)

type FileStorage struct {
	StorageStruct
}

func NewFileStorager(P *config.Param) Storager {
	ReadStorage(P)
	return &FileStorage{
		StorageStruct: StorageStruct{
			UserID: "",
			Key:    "",
			Value:  "",
		},
	}
}

func (s *FileStorage) SetShortURL(fURL, UserID string, Params *config.Param) (string, error) {
	s.Key = HashStr(fURL)
	_, true := BaseURL[s.Key]
	if true {
		if UserURL[s.Key] == UserID {
			return s.Key, ErrConflict
		}
	}

	var mutex sync.Mutex
	mutex.Lock()
	BaseURL[s.Key] = fURL
	UserURL[s.Key] = UserID
	DeletedURL[s.Key] = false
	mutex.Unlock()
	file, err := NewWriterFile(Params)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteFile(s.Key, UserID, fURL)
	return s.Key, nil
}

func (s *FileStorage) RetFullURL(key string) (string, error) {
	var mutex sync.RWMutex
	mutex.Lock()
	del := DeletedURL[s.Key]
	mutex.Unlock()
	if del {
		return "", ErrGone
	}
	return BaseURL[key], nil
}

type readerFile struct {
	file    *os.File
	decoder *json.Decoder
}

func (s *FileStorage) ReturnAllURLs(UserID string, P *config.Param) ([]byte, error) {
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

func (s *FileStorage) CheckPing(P *config.Param) error {
	return errors.New("wrong DB used: file storage")
}

func (s *FileStorage) WriteMultiURL(m *[]MultiURL, UserID string, P *config.Param) (*[]MultiURL, error) {
	r := make([]MultiURL, len(*m))
	file, err := NewWriterFile(P)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for i, v := range *m {
		Key := HashStr(v.OriginURL)
		var mutex sync.RWMutex
		mutex.Lock()
		BaseURL[Key] = v.OriginURL
		UserURL[Key] = UserID
		DeletedURL[s.Key] = false
		mutex.Unlock()
		file.WriteFile(s.Key, UserID, v.OriginURL)
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(P.URL + "/" + Key)
	}

	return &r, nil
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
		log.Println(err)
	}
}

func (w *writerFile) Close() error {
	return w.file.Close()
}

func (s *FileStorage) CloseDB() {
	log.Println("file closed")
}

func (s *FileStorage) MarkDeleted(DeleteURLs *[]string, UserID string, P *config.Param) {
	inputCh := make(chan string)
	go func() {
		for _, del := range *DeleteURLs {
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
		DeletedURL[key] = true
	}
}

func (s *FileStorage) newWorker(in, out chan string, UserID string) {
	go func() {
		for myURL := range in {
			key := strings.Trim(myURL, "\"")
			var mutex sync.RWMutex
			mutex.Lock()
			id := UserURL[s.Key]
			mutex.Unlock()
			if id != UserID {
				continue
			}
			out <- key
		}
		close(out)
	}()
}
