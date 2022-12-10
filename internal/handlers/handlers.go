package handlers

import (
	"io"
	"net/http"
	"net/url"
	"shortURL/internal/keygen"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	baseURL = make(map[string]string)
	key     string
)

func NewRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fURL := string(bytes)
		_, err = url.Parse(fURL)
		if err != nil {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
			return
		}
		key := SetShortURL(fURL)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://" + r.Host + "/" + key))
	})
	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "id")
		address := RetFullURL(key)
		if address == "" {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
		}
		http.Redirect(w, r, address, http.StatusTemporaryRedirect)
	})
	return r
}

func SetShortURL(fURL string) string {

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

	baseURL[key] = fURL
	return key
}

func RetFullURL(key string) string {
	return baseURL[key]
}
