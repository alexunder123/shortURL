package router

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"

	"shortURL/internal/storage"
)

func (m Router) BatchPost(w http.ResponseWriter, r *http.Request) {
	userID := ReadContextID(r)
	if userID == "" {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	var multiURLs = make([]storage.MultiURL, 0)
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("BatchPost read body err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(bytes, &multiURLs); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	if len(multiURLs) == 0 {
		http.Error(w, "batch URLs empty", http.StatusNoContent)
		return
	}
	rMultiURLs, err := m.str.WriteMultiURL(multiURLs, userID, m.prm)
	if err != nil {
		log.Error().Err(err).Msg("BatchPost WriteMultiURL err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rMultiURLsBZ, err := json.Marshal(rMultiURLs)
	if err != nil {
		log.Error().Err(err).Msg("BatchPost json.Marshal err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write(rMultiURLsBZ)
}

func (m Router) ShortenPost(w http.ResponseWriter, r *http.Request) {
	userID := ReadContextID(r)
	if userID == "" {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	var addr postURL
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("ShortenPost read body err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(bytes, &addr); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	_, err = url.Parse(addr.GetURL)
	if err != nil {
		log.Error().Err(err).Msg("ShortenPost url.Parse err")
		http.Error(w, "Wrong address!", http.StatusBadRequest)
		return
	}

	key, err := m.str.SetShortURL(addr.GetURL, userID, m.prm)
	if errors.Is(err, storage.ErrConflict) {
		newAddr := postURL{SetURL: m.prm.URL + "/" + key}
		newAddrBZ, err := json.Marshal(newAddr)
		if err != nil {
			log.Error().Err(err).Msg("ShortenPost json.Marshal err")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write(newAddrBZ)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("ShortenPost SetShortURL err")
		http.Error(w, "Wrong address!", http.StatusInternalServerError)
		return
	}
	newAddr := postURL{SetURL: m.prm.URL + "/" + key}
	newAddrBZ, err := json.Marshal(newAddr)
	if err != nil {
		log.Error().Err(err).Msg("ShortenPost json.Marshal err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write(newAddrBZ)
}

func (m Router) URLPost(w http.ResponseWriter, r *http.Request) {
	userID := ReadContextID(r)
	if userID == "" {
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
	key, err := m.str.SetShortURL(fURL, userID, m.prm)
	if errors.Is(err, storage.ErrConflict) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(m.prm.URL + "/" + key))
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("URLPost storage err")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(m.prm.URL + "/" + key))
}
