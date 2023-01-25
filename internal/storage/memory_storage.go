package storage

import (
	"encoding/json"
	"errors"
	"shortURL/internal/config"
	"strings"
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

	BaseURL[key] = fURL
	UserURL[key] = UserID
	return key, nil
}

func (s *MemoryStorage) RetFullURL(key string) string {
	s.RLock()
	del := s.DeletedURL[key]
	s.RUnlock()
	if del {
		return "", ErrGone
	}
	return s.baseURL[key], nil
}

func (s *MemoryStorage) ReturnAllURLs(UserID string, P *config.Param) ([]byte, error) {
	if len(BaseURL) == 0 {
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
	Mutex.Unlock()
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
