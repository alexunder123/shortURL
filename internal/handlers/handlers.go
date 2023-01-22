package handlers

import (
	"shortURL/internal/config"
	"shortURL/internal/storage"
	"shortURL/internal/router"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(P *config.Param, S storage.Storager) *router.Router {
	r := router.Router{P: P, S: S}
	r.Router = chi.NewRouter()

	r.Router.Use(middleware.Logger)
	r.Router.Use(middleware.Recoverer)
	r.Router.Use(router.MidWareDecompress)
	r.Router.Use(router.MidWareCookies)

	r.Router.Post("/api/shorten/batch", r.BatchPost)
	r.Router.Post("/api/shorten", r.ShortenPost)
	r.Router.Post("/", r.URLPost)

	r.Router.Get("/api/user/urls", r.URLsGet)
	r.Router.Get("/{id}", r.IdGet)
	r.Router.Get("/ping", r.PingGet)

	return &r
}
