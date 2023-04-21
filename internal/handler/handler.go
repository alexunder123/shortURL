package handler

import (
	"net"
	"shortURL/internal/config"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

// Handler хранит ссылки на параметры сервиса.
type Handler struct {
	cfg       *config.Config
	strg      storage.Storager
	workerDel *worker.Worker
	Subnet    net.IPNet
}

// NewHandler генерирует структуру Handler.
func NewHandler(cfg *config.Config, strg storage.Storager, wrkr *worker.Worker) *Handler {
	h := Handler{
		cfg:       cfg,
		strg:      strg,
		workerDel: wrkr,
	}
	if cfg.TrustedSubnet != "" {
		_, subnet, _ := net.ParseCIDR(cfg.TrustedSubnet)
		h.Subnet = *subnet
	}
	return &h
}

type postURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}
