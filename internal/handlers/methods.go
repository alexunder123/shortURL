package handlers

import (
	"log"
	"math/rand"
	"shortURL/internal/app"
	"sync"
	"time"
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
		key = RandomStr()
		_, err := baseURL[key]
		if !err {
			break
		}
	}

	var mutex sync.Mutex
	mutex.Lock()
	baseURL[key] = fURL
	if Params.SaveDB {
		file, err := app.NewWriterDB(Params)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		file.WriteDB(key, fURL)
	}
	mutex.Unlock()
	return key
}

func RetFullURL(key string) string {
	return baseURL[key]
}

func RandomStr() string {
	const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	n := 6
	bts := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		bts[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(bts)
}
