package router

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"shortURL/internal/storage"

	"github.com/rs/zerolog/log"
)

func (m Router) BatchPost(w http.ResponseWriter, r *http.Request) {
	userID := ReadContextID(r)
	if userID == "" {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	var MultiURLs = make([]storage.MultiURL, 0)
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(bytes, &MultiURLs); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	RMultiURLs, err := m.Str.WriteMultiURL(MultiURLs, userID, m.Prm)
	if err != nil {
		log.Error().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	RMultiURLsBZ, err := json.Marshal(RMultiURLs)
	if err != nil {
		log.Error().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write(RMultiURLsBZ)
}

func (m Router) ShortenPost(w http.ResponseWriter, r *http.Request) {
	userID := ReadContextID(r)
	if userID == "" {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	var addr PostURL
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(bytes, &addr); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	_, err = url.Parse(addr.GetURL)
	if err != nil {
		log.Error().Err(err)
		http.Error(w, "Wrong address!", http.StatusBadRequest)
		return
	}

	key, err := m.Str.SetShortURL(addr.GetURL, userID, m.Prm)
	if err == storage.ErrConflict {
		newAddr := PostURL{SetURL: m.Prm.URL + "/" + key}
		newAddrBZ, err := json.Marshal(newAddr)
		if err != nil {
			log.Error().Err(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write(newAddrBZ)
		return
	}
	newAddr := PostURL{SetURL: m.Prm.URL + "/" + key}
	newAddrBZ, err := json.Marshal(newAddr)
	if err != nil {
		log.Error().Err(err)
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
		log.Error().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fURL := string(bytes)
	_, err = url.Parse(fURL)
	if err != nil {
		log.Error().Err(err)
		http.Error(w, "Wrong address!", http.StatusBadRequest)
		return
	}
	key, err := m.Str.SetShortURL(fURL, userID, m.Prm)
	if err == storage.ErrConflict {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(m.Prm.URL + "/" + key))
		return
	}
	if err != nil {
		log.Error().Err(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(m.Prm.URL + "/" + key))
}
