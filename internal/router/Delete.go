package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (m Router) URLsDelete(w http.ResponseWriter, r *http.Request) {
	urlsBZ, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("URLsDelete read body error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var deleteURLs []string
	if err = json.Unmarshal(urlsBZ, &deleteURLs); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	userID := ReadContextID(r)
	log.Debug().Msgf("Received URLs to delete: %s", deleteURLs)
	go m.str.MarkDeleted(deleteURLs, userID)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write(nil)
}
