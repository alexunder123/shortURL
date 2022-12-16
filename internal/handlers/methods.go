package handlers

import (
	"log"
	"shortURL/internal/app"
	"shortURL/internal/keygen"
	"sync"
)

var (
	baseURL = make(map[string]string)
	key     string
)

func SetShortURL(fURL string, Params *app.Param) string {
	for key, addr := range baseURL {
		if addr == fURL {
			return key
		}
	}
	for {
		key = keygen.RandomStr()
		_, err := baseURL[key]
		if !err {
			break
		}
	}

	var mutex sync.Mutex
	mutex.Lock()
	baseURL[key] = fURL
	file, err := app.NewWriterDB(Params)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteDB(key, fURL)
	mutex.Unlock()
	return key
}

func RetFullURL(key string) string {
	return baseURL[key]
}
