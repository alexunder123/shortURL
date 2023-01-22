package router

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"shortURL/internal/storage"
)

func (m Router) BatchPost(w http.ResponseWriter, r *http.Request) {
	var MultiURLs = make([]storage.MultiURL, 0)
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(bytes, &MultiURLs); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	userID := ReadContextID(r)
	RMultiURLs, err := m.S.WriteMultiURL(&MultiURLs, userID, m.P)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	RMultiURLsBZ, err := json.Marshal(RMultiURLs)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write(RMultiURLsBZ)
}

func (m Router) ShortenPost(w http.ResponseWriter, r *http.Request) {
	var Addr PostURL
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(bytes, &Addr); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	_, err = url.Parse(Addr.GetURL)
	if err != nil {
		log.Println(err)
		http.Error(w, "Wrong address!", http.StatusBadRequest)
		return
	}
	userID := ReadContextID(r)

	key, err := m.S.SetShortURL(Addr.GetURL, userID, m.P)
	if err == storage.ErrConflict {
		NewAddr := PostURL{SetURL: m.P.URL + "/" + key}
		NewAddrBZ, err := json.Marshal(NewAddr)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write(NewAddrBZ)
		return
	}
	NewAddr := PostURL{SetURL: m.P.URL + "/" + key}
	NewAddrBZ, err := json.Marshal(NewAddr)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write(NewAddrBZ)
}

func (m Router) URLPost(w http.ResponseWriter, r *http.Request) {
	userID := ReadContextID(r)
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fURL := string(bytes)
	_, err = url.Parse(fURL)
	if err != nil {
		log.Println(err)
		http.Error(w, "Wrong address!", http.StatusBadRequest)
		return
	}
	key, err := m.S.SetShortURL(fURL, userID, m.P)
	if err == storage.ErrConflict {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(m.P.URL + "/" + key))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(m.P.URL + "/" + key))
}
