package router

import (
	"net/http"
	"shortURL/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (m Router) URLsGet(w http.ResponseWriter, r *http.Request) {
	userID := ReadContextID(r)
	if userID == "" {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	urlsBZ, err := m.Str.ReturnAllURLs(userID, m.Prm)
	if err == storage.ErrNoContent {
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

func (m Router) IDGet(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "id")
	address, err := m.Str.RetFullURL(key)
	if err == storage.ErrGone {
		http.Error(w, "URL Deleted", http.StatusGone)
	}
	if err != nil {
		log.Error().Err(err).Msg("IDGet storage error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if address == "" {
		http.Error(w, "Wrong address!", http.StatusBadRequest)
	}
	http.Redirect(w, r, address, http.StatusTemporaryRedirect)
}

func (m Router) PingGet(w http.ResponseWriter, r *http.Request) {
	err := m.Str.CheckPing(m.Prm)
	if err != nil {
		log.Error().Err(err).Msg("PingGet DB error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
