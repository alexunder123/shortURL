package router

import (
	"shortURL/internal/config"
	"shortURL/internal/storage"
)

type Router struct {
	Prm *config.Param
	Str storage.Storager
}

func NewRouter(P *config.Param, S storage.Storager) *Router {
	return &Router{
		Prm: P,
		Str: S,
	}
}

type PostURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}
