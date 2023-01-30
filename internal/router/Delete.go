package router

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

func (m Router) URLsDelete(w http.ResponseWriter, r *http.Request) {
	urlsBZ, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("URLsDelete read body error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urlsBZ = bytes.Trim(urlsBZ, "[]")
	urls := string(urlsBZ)
	deleteURLs := strings.Split(urls, ",")
	userID := ReadContextID(r)
	log.Info().Msgf("Received URLs to delete:", deleteURLs)
	go m.Str.MarkDeleted(deleteURLs, userID, m.Prm)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write(nil)
}
