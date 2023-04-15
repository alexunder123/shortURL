package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"shortURL/internal/midware"
	"shortURL/internal/storage"
)

// URLsDelete метод обрабатывает запрос на удаление записей сокращенных адресов.
func (h *Handler) URLsDelete(w http.ResponseWriter, r *http.Request) {
	urlsBZ, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("URLsDelete read body error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userID, ok := r.Context().Value(midware.UserID).(string)
	if !ok {
		http.Error(w, "userID empty", http.StatusUnauthorized)
		return
	}
	err = h.workerDel.Add(urlsBZ, userID)
	if errors.Is(err, storage.ErrUnsupported) {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	if errors.Is(err, storage.ErrUnavailable) {
		http.Error(w, "the server is in the process of stopping", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write(nil)
}
