package handlers

import (
	"shortURL/internal/router"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewHandler(r *router.Router) *chi.Mux {
	h := chi.NewRouter()

	h.Use(middleware.Logger)
	h.Use(middleware.Recoverer)
	h.Use(router.MidWareDecompress)
	h.Use(router.MidWareCookies)

	h.Post("/api/shorten/batch", r.BatchPost)
	h.Post("/api/shorten", r.ShortenPost)
	h.Post("/", r.URLPost)

	h.Get("/api/user/urls", r.URLsGet)
	h.Get("/{id}", r.IDGet)
	h.Get("/ping", r.PingGet)

	return h
}
