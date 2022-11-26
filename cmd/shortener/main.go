package main

import (
	//_ "crypto/MD5"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var (
	BaseURL = make(map[string]string)
)

func ShortenerURL(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		Url, err := io.ReadAll(r.Body)
		fURL := string(Url)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// _, err := url.Parse(Url)
		// if err != nil {
		// 	http.Error(w, "Wrong address!", http.StatusBadRequest)
		// 	return
		// }
		for key, addr := range BaseURL {
			if addr == fURL {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(key))
				return
			}
		}
		key := RandomStr()
		BaseURL[key] = fURL
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + key))
	case http.MethodGet:
		key := r.URL.Path
		key = strings.TrimPrefix(key, "/")
		// _, err := url.Parse(Url)
		// if err != nil {
		// 	http.Error(w, "Wrong address!", http.StatusBadRequest)
		// 	return
		// }
		addr, isExist := BaseURL[key]
		if !isExist {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, addr, 307)
	default:
		http.Error(w, "Wrong request!", http.StatusBadRequest)
		return
	}

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

func main() {

	http.HandleFunc("/", ShortenerURL)
	http.ListenAndServe("localhost:8080", nil)
}
