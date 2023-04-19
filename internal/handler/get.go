package handler

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"shortURL/internal/midware"
	"shortURL/internal/storage"
)

// URLsGet метод возвращает пользователю список сокращенных им адресов.
func (h *Handler) URLsGet(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(midware.UserID).(string)
	if !ok {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	urls, err := h.strg.ReturnAllURLs(userID, h.cfg)
	if errors.Is(err, storage.ErrNoContent) {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	urlsBZ, err := json.Marshal(urls)
	if err != nil {
		log.Error().Err(err).Msg("URLsGet json.Marshal error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(urlsBZ)
}

// IDGet метод возвращает пользователю исходный адрес.
func (h *Handler) IDGet(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "id")
	address, err := h.strg.RetFullURL(key)
	if errors.Is(err, storage.ErrGone) {
		http.Error(w, "URL Deleted", http.StatusGone)
		return
	}
	if errors.Is(err, storage.ErrNoContent) {
		http.Error(w, "Wrong address!", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("IDGet storage error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, address, http.StatusTemporaryRedirect)
}

// PingGet метод возвращает статус наличия соединения с базой данных.
func (h *Handler) PingGet(w http.ResponseWriter, r *http.Request) {
	err := h.strg.CheckPing(h.cfg)
	if err != nil {
		log.Error().Err(err).Msg("PingGet DB error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// StatsGet метод возвращает количество сокращенных URL и пользователей в сервисе.
func (h *Handler) StatsGet(w http.ResponseWriter, r *http.Request) {
	if h.cfg.TrustedSubnet == "" {
		log.Error().Msgf("StatsGet TrustedSubnet isn't determined")
		http.Error(w, "TrustedSubnet isn't determined", http.StatusForbidden)
		return
	}
	userIP := net.ParseIP(r.Header.Get("X-Real-IP"))
	if userIP == nil {
		log.Error().Msgf("StatsGet User IP-address not resolved")
		http.Error(w, "User IP-address not resolved", http.StatusBadRequest)
		return
	}
	if !h.Subnet.Contains(userIP) {
		log.Error().Msgf("StatsGet User IP-address isn't CIDR subnet")
		http.Error(w, "User IP-address isn't CIDR subnet", http.StatusForbidden)
		return
	}

	stats, err := h.strg.ReturnStats()
	if err != nil {
		log.Error().Err(err).Msg("StatsGet ReturnStats error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	statsBZ, err := json.Marshal(stats)
	if err != nil {
		log.Error().Err(err).Msg("StatsGet json.Marshal error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(statsBZ)
}
