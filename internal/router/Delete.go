package router

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Заменить на Handler
func (r *Handler) URLsDelete(w http.ResponseWriter, r *http.Request) {
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
	if m.closed {
		http.Error(w, "the server is in the process of stopping", http.StatusServiceUnavailable)
		return
	}

	// FanIn (канал, который передаёт в воркер на удаление)
	r.DelChan <- { userID, deleteURLs }

	//go m.writeToDelInput(deleteURLs, userID)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write(nil)
}

// ! Отдельный пакет: воркер удаления, FanIn
func (r Handler) ProcessingDel(ctx context.Context) context.Context {
	ctxStop, cancelStop := context.WithCancel(context.Background())
	go func() {
		//update
		log.Debug().Msg("ProcessingDel started")
		for toDel := range m.inputCh {
			m.str.MarkDeleted(toDel.Keys, toDel.ID)
		}
		log.Info().Msg("output finished")
		cancelStop()
	}()

	go func() {
		<-ctx.Done()
		m.closed = true
		close(m.inputCh)
		log.Info().Msg("inputCh closed")
	}()

	return ctxStop
}

func (m Router) writeToDelInput(deleteURLs []string, userID string) {
	m.inputCh <- ToDelete{Keys: deleteURLs, ID: userID}
}
