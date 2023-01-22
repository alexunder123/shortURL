package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"shortURL/internal/config"
	"shortURL/internal/storage"
	"strings"
)

func URLsDelete(w http.ResponseWriter, r *http.Request, P *config.Param, S storage.Storager) {
	URLsBZ, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	URLsBZ = bytes.TrimLeft(URLsBZ, "[")
	URLsBZ = bytes.TrimRight(URLsBZ, "]")
	URLs := string(URLsBZ)
	DeleteURLs := strings.Split(URLs, ",")
	UserID := ReadContextID(r)
	go s.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write(nil)
}
