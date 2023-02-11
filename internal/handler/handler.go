package handler

import (
	"shortURL/internal/config"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

type Handler struct {
	cfg       *config.Config
	strg      storage.Storager
	workerDel *worker.Worker
}

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
