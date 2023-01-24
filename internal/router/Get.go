package router

import (
	"log"
	"net/http"
	"shortURL/internal/storage"

	"github.com/go-chi/chi/v5"
)

func (m Router) URLsGet(w http.ResponseWriter, r *http.Request) {
	userID := ReadContextID(r)
	URLsBZ, err := m.S.ReturnAllURLs(userID, m.P)
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

func (m Router) IDGet(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "id")
	address, err := m.S.RetFullURL(key)
	if err == storage.ErrGone {
		http.Error(w, "URL Deleted", http.StatusGone)
	}
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if address == "" {
		http.Error(w, "Wrong address!", http.StatusBadRequest)
	}
	http.Redirect(w, r, address, http.StatusTemporaryRedirect)
}

func (m Router) PingGet(w http.ResponseWriter, r *http.Request) {
	err := m.S.CheckPing(m.P)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
