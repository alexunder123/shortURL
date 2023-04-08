// Модуль вызывает обработчик соответствующий параметрам API запроса.
package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"shortURL/internal/handler"
	"shortURL/internal/midware"
)

// NewRouter функция генерирует мультиплексор для обработки API запросов.
func NewRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(midware.Decompress)
	r.Use(midware.Cookies)

	r.Post("/api/shorten/batch", h.BatchNewEtriesPost)
	r.Post("/api/shorten", h.ShortenPost)
	r.Post("/", h.URLPost)

	r.Get("/api/user/urls", h.URLsGet)
	r.Get("/api/internal/stats", h.StatsGet)
	r.Get("/{id}", h.IDGet)
	r.Get("/ping", h.PingGet)

	r.Delete("/api/user/urls", h.URLsDelete)

	return r
}
