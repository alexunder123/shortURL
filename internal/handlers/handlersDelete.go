package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
)

func (m Router) URLsDelete(w http.ResponseWriter, r *http.Request) {
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
	userID := ReadContextID(r)
	go m.S.MarkDeleted(&DeleteURLs, userID, m.P)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write(nil)
}
