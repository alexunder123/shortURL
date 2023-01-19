package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"shortURL/internal/config"
	"shortURL/internal/storage"
)

func batchPost(w http.ResponseWriter, r *http.Request, P *config.Param, S storage.Storager) {
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
	UserID := ReadContextID(r)
	RMultiURLs, err := S.WriteMultiURL(&MultiURLs, UserID, P)
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

func shortenPost(w http.ResponseWriter, r *http.Request, P *config.Param, S storage.Storager) {
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
	UserID := ReadContextID(r)

	key, err := S.SetShortURL(Addr.GetURL, UserID, P)
	if err == storage.ErrConflict {
		NewAddr := PostURL{SetURL: P.URL + "/" + key}
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
	NewAddr := PostURL{SetURL: P.URL + "/" + key}
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

func URLPost(w http.ResponseWriter, r *http.Request, P *config.Param, S storage.Storager) {
	UserID := ReadContextID(r)
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
	key, err := S.SetShortURL(fURL, UserID, P)
	if err == storage.ErrConflict {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(P.URL + "/" + key))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(P.URL + "/" + key))
}
