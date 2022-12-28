package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"shortURL/internal/config"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5"
)

type SQLStorage struct {
	StorageStruct
}

func NewSQLStorager() Storager {
	return &SQLStorage{
		StorageStruct: StorageStruct{
			UserID: "",
			Key:    "",
			Value:  "",
		},
	}
}

func (s *SQLStorage) SetShortURL(fURL, UserID string, Params *config.Param) string {
	s.Key = HashStr(fURL)
	_, true := BaseURL[s.Key]
	if true {
		return s.Key
	}

	var mutex sync.RWMutex
	mutex.Lock()
	BaseURL[s.Key] = fURL
	UserURL[s.Key] = UserID
	mutex.Unlock()
	return s.Key
}

func (s *SQLStorage) RetFullURL(key string) string {
	return BaseURL[key]
}

func (s *SQLStorage) ReturnAllURLs(UserID string, P *config.Param) ([]byte, error) {
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

func (s *SQLStorage) CheckPing(P *config.Param) error {
	db, err := sql.Open("PostgreSQL", P.SQL)
	if err != nil {
		return err
	}
	defer db.Close()
	var ctx context.Context
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	return err
}
