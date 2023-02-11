package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"shortURL/internal/midware"
	"shortURL/internal/worker"
)

func (h *Handler) URLsDelete(w http.ResponseWriter, r *http.Request) {
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
	userID, ok := r.Context().Value(midware.UserID).(string)
	if !ok {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	log.Debug().Msgf("Received URLs to delete: %s", deleteURLs)
	log.Debug().Msgf("Received ID to delete: %s", userID)
	if h.workerDel.Closed {
		http.Error(w, "the server is in the process of stopping", http.StatusServiceUnavailable)
		return
	}
	h.workerDel.InputCh <- worker.ToDelete{Keys: deleteURLs, ID: userID}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write(nil)
}
