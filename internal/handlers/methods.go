package handlers

import (
	"log"
	"math/rand"
	"shortURL/internal/app"
	"sync"
	"time"
)

var (
	key string
)

func SetShortURL(fURL string, Params *app.Param) string {
	for key, addr := range app.BaseURL {
		if addr == fURL {
			return key
		}
	}
	for {
		key = RandomStr()
		_, err := app.BaseURL[key]
		if !err {
			break
		}
	}

	var mutex sync.Mutex
	mutex.Lock()
	app.BaseURL[key] = fURL
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
	return app.BaseURL[key]
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
