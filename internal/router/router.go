package router

import (
	"shortURL/internal/config"
	"shortURL/internal/storage"
)

type Router struct {
	prm     *config.Param
	str     storage.Storager
}

func NewRouter(P *config.Param, S storage.Storager) *Router {
	return &Router{
		prm:     P,
		str:     S,
	}
}

type postURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}
