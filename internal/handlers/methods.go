package handlers

import (
	"compress/gzip"
	"fmt"

	"crypto/md5"
	"io"
	"log"
	"math/rand"
	"net/http"
	"shortURL/internal/app"
	"strings"
	"sync"
	"time"
)

var (
	key string
)

func SetShortURL(fURL string, Params *app.Param) string {
	key = HashStr(fURL)
	_, true := app.BaseURL[key]
	if true {
		return key
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

func HashStr(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			decompressed := gz
			defer gz.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(decompressed)
			next.ServeHTTP(w, r)
			return
		} else if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
			return
		} else {
			next.ServeHTTP(w, r)
			return
		}
	})
}
