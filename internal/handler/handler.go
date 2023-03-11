package handler

import (
	"shortURL/internal/config"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

// Handler хранит ссылки на параметры сервиса.
type Handler struct {
	cfg       *config.Config
	strg      storage.Storager
	workerDel *worker.Worker
}

// NewHandler генерирует структуру Handler.
func NewHandler(cfg *config.Config, strg storage.Storager, wrkr *worker.Worker) *Handler {
	return &Handler{
		cfg:       cfg,
		strg:      strg,
		workerDel: wrkr,
	}
}

type postURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}
