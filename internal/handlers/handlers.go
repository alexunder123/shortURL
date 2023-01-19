package handlers

import (
	"net/http"
	"shortURL/internal/config"
	"shortURL/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type PostURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}

func NewRouter(P *config.Param, S storage.Storager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(Decompress)
	r.Use(Cookies)

	r.Post("/api/shorten/batch", func(w http.ResponseWriter, r *http.Request) {
		batchPost(w, r, P, S)
	})

	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		shortenPost(w, r, P, S)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		URLPost(w, r, P, S)
	})

	r.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		URLsGet(w, r, P, S)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		idGet(w, r, P, S)
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		pingGet(w, r, P, S)
	})

	return r
}
