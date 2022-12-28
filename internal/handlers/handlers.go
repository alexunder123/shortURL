package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"shortURL/internal/config"
	"shortURL/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type PostURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}

func NewRouter(P *config.Param, S storage.Storager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(Decompress)
	r.Use(Cookies)

	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		var Addr PostURL
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = json.Unmarshal(bytes, &Addr); err != nil {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
			return
		}
		_, err = url.Parse(Addr.GetURL)
		if err != nil {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
			return
		}
		UserID := ReadContextID(r)

		key := S.SetShortURL(Addr.GetURL, UserID, P)
		NewAddr := PostURL{SetURL: P.URL + "/" + key}
		NewAddrBZ, err := json.Marshal(NewAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(NewAddrBZ)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		UserID := ReadContextID(r)
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fURL := string(bytes)
		_, err = url.Parse(fURL)
		if err != nil {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
			return
		}
		key := S.SetShortURL(fURL, UserID, P)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(P.URL + "/" + key))
	})

	r.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
		UserID := ReadContextID(r)
		URLsBZ, err := S.ReturnAllURLs(UserID, P)
		if err == storage.ErrNoContent {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(URLsBZ)
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "id")
		address := S.RetFullURL(key)
		if address == "" {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
		}
		http.Redirect(w, r, address, http.StatusTemporaryRedirect)
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		err := S.CheckPing(P)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	})

	return r
}
