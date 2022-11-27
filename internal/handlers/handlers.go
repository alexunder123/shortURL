package handlers

import (
	"io"
	"net/http"
	"net/url"
	"shortURL/internal/keygen"
	"strings"
)

var (
	BaseURL = make(map[string]string)
)

func ShortenerURL(w http.ResponseWriter, r *http.Request) {
	var key string
	switch r.Method {
	case http.MethodPost:
		Url, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fURL := string(Url)
		_, err = url.Parse(fURL)
		if err != nil {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
			return
		}
		for key, addr := range BaseURL {
			if addr == fURL {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(key))
				return
			}
		}
		for {
			key = keygen.RandomStr()
			_, err := BaseURL[key]
			if !err {
				break
			}
		}

		BaseURL[key] = fURL
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + key))
	case http.MethodGet:
		key = r.URL.Path
		_, err := url.Parse(key)
		if err != nil {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
			return
		}
		key = strings.TrimPrefix(key, "/")
		addr, isExist := BaseURL[key]
		if !isExist {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, addr, http.StatusTemporaryRedirect)
	default:
		http.Error(w, "Wrong request!", http.StatusBadRequest)
		return
	}

}
