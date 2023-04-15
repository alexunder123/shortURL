package handler

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"

	"shortURL/internal/midware"
	"shortURL/internal/storage"
)

// BatchNewEtriesPost метод принимает от пользователя и возвращает JSON список адресов на сокращение.
func (h *Handler) BatchNewEtriesPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(midware.UserID).(string)
	if !ok {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("BatchPost read body err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rMultiURLsBZ, err := h.strg.WriteMultiURL(bytes, userID, h.cfg)
	if errors.Is(err, storage.ErrUnsupported) {
		log.Error().Err(err).Msg("WriteMultiURL json error")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	if errors.Is(err, storage.ErrNoContent) {
		log.Error().Err(err).Msg("WriteMultiURL json no content")
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("BatchPost WriteMultiURL err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write(rMultiURLsBZ)
}

// ShortenPost метод принимает от пользователя и возвращает в JSON адрес на сокращение.
func (h *Handler) ShortenPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(midware.UserID).(string)
	if !ok {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("ShortenPost read body err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newAddrBZ, err := h.strg.SetShortURLjs(bytes, userID, h.cfg)
	if errors.Is(err, storage.ErrBadRequest) {
		log.Error().Err(err).Msg("ShortenPost url.Parse err")
		http.Error(w, "Wrong address!", http.StatusBadRequest)
		return
	}
	if errors.Is(err, storage.ErrUnsupported) {
		log.Error().Err(err).Msg("SetShortURL json error")
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	if errors.Is(err, storage.ErrConflict) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write(newAddrBZ)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("ShortenPost SetShortURL err")
		http.Error(w, "ShortenPost json.Marshal err", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write(newAddrBZ)
}

// URLPost метод принимает от пользователя и возвращает адрес на сокращение.
func (h *Handler) URLPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(midware.UserID).(string)
	if !ok {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("URLPost read body err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fURL := string(bytes)
	_, err = url.Parse(fURL)
	if err != nil {
		log.Error().Err(err).Msg("URLPost url.Parse err")
		http.Error(w, "Wrong address!", http.StatusBadRequest)
		return
	}
	newAddr, err := h.strg.SetShortURL(fURL, userID, h.cfg)
	if errors.Is(err, storage.ErrConflict) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(newAddr))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("URLPost storage err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(newAddr))
}
