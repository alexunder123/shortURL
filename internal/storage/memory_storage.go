package storage

import (
	"encoding/json"
	"errors"
	"log"
	"shortURL/internal/config"
	"strings"
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

func (s *MemoryStorage) SetShortURL(fURL, UserID string, Params *config.Param) (string, error) {
	s.Key = HashStr(fURL)
	_, true := BaseURL[s.Key]
	if true {
		if UserURL[s.Key] == UserID {
			return s.Key, ErrConflict
		}
	}

	var mutex sync.RWMutex
	mutex.Lock()
	BaseURL[s.Key] = fURL
	UserURL[s.Key] = UserID
	mutex.Unlock()
	return s.Key, nil
}

func (s *MemoryStorage) RetFullURL(key string) (string, error) {
	var mutex sync.RWMutex
	mutex.Lock()
	if DeletedURL[s.Key] {
		return "", ErrGone
	}
	mutex.Unlock()
	return BaseURL[key], nil
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

func (s *MemoryStorage) CheckPing(P *config.Param) error {
	return errors.New("wrong DB used: memory storage")
}

func (s *MemoryStorage) WriteMultiURL(m *[]MultiURL, UserID string, P *config.Param) (*[]MultiURL, error) {
	r := make([]MultiURL, len(*m))
	for i, v := range *m {
		Key := HashStr(v.OriginURL)
		var mutex sync.RWMutex
		mutex.Lock()
		BaseURL[Key] = v.OriginURL
		UserURL[Key] = UserID
		mutex.Unlock()
		r[i].CorrID = v.CorrID
		r[i].ShortURL = string(P.URL + "/" + Key)
	}

	return &r, nil
}

func (s *MemoryStorage) CloseDB() {
	log.Println("closed")
}

func (s *MemoryStorage) MarkDeleted(DeleteURLs *[]string, UserID string, P *config.Param) {
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
	var mutex sync.Mutex
	mutex.Lock()
	for key := range outCh {
		DeletedURL[key] = true
	}
	mutex.Unlock()
}

func (s *MemoryStorage) newWorker(in, out chan string, UserID string) {
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
