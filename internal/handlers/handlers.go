package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"shortURL/internal/router"
)

// Router == Mux == Multiplexer
func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(router.MidWareDecompress)
	r.Use(router.MidWareCookies)

	r.Post("/api/shorten/batch", h.BatchPost)
	r.Post("/api/shorten", h.ShortenPost)
	r.Post("/", r.URLPost)

	r.Get("/api/user/urls", h.URLsGet)
	r.Get("/{id}", h.IDGet)
	r.Get("/ping", h.PingGet)

	r.Delete("/api/user/urls", h.URLsDelete)

	return r
}
