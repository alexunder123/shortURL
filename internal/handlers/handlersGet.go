package handlers

import (
	"log"
	"net/http"
	"shortURL/internal/config"
	"shortURL/internal/storage"

	"github.com/go-chi/chi/v5"
)

func URLsGet(w http.ResponseWriter, r *http.Request, P *config.Param, S storage.Storager) {
	UserID := ReadContextID(r)
	URLsBZ, err := S.ReturnAllURLs(UserID, P)
	if err == storage.ErrNoContent {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(URLsBZ)
}

func idGet(w http.ResponseWriter, r *http.Request, P *config.Param, S storage.Storager) {
	key := chi.URLParam(r, "id")
	address := S.RetFullURL(key)
	if address == "" {
		http.Error(w, "Wrong address!", http.StatusBadRequest)
	}
	http.Redirect(w, r, address, http.StatusTemporaryRedirect)
}

func pingGet(w http.ResponseWriter, r *http.Request, P *config.Param, S storage.Storager) {
	err := S.CheckPing(P)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
