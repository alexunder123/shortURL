package router

import (
	"shortURL/internal/config"
	"shortURL/internal/storage"
)

type Router struct {
	prm     *config.Param
	str     storage.Storager
	inputCh chan ToDelete
	closed bool
}

func NewRouter(P *config.Param, S storage.Storager) *Router {
	return &Router{
		prm:     P,
		str:     S,
		inputCh: make(chan ToDelete),
		closed: false,
	}
}

type postURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}


type ToDelete struct{
	Keys []string
	ID string
}