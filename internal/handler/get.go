package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"shortURL/internal/midware"
	"shortURL/internal/storage"
)

func (h *Handler) URLsGet(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(midware.UserID).(string)
	if !ok {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	urlsBZ, err := h.strg.ReturnAllURLs(userID, h.cfg)
	if errors.Is(err, storage.ErrNoContent) {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	if err != nil {
		log.Error().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(urlsBZ)
}

func (h *Handler) IDGet(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "id")
	address, err := h.strg.RetFullURL(key)
	if errors.Is(err, storage.ErrGone) {
		http.Error(w, "URL Deleted", http.StatusGone)
		return
	}
	if errors.Is(err, storage.ErrNoContent) {
		http.Error(w, "Wrong address!", http.StatusBadRequest)
	}
	if err != nil {
		log.Error().Err(err).Msg("IDGet storage error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, address, http.StatusTemporaryRedirect)
}

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
