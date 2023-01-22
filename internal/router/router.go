package router

import (
	"shortURL/internal/config"
	"shortURL/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	Router *chi.Mux
	P      *config.Param
	S      storage.Storager
}

type PostURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}
