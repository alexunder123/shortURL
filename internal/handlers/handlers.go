package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"shortURL/internal/app"
	"shortURL/internal/keygen"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	baseURL = make(map[string]string)
	key     string
)

type PostURL struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}

func NewRouter(Params *app.Param) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/api/shorten", func(w http.ResponseWriter, r *http.Request) {
		var Addr PostURL
		bytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()
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
		key := SetShortURL(Addr.GetURL)
		NewAddr := PostURL{SetURL: Params.URL + "/" + key}
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
		bytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()
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
		key := SetShortURL(fURL)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(Params.URL + "/" + key))
	})

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "id")
		address := RetFullURL(key)
		if address == "" {
			http.Error(w, "Wrong address!", http.StatusBadRequest)
		}
		http.Redirect(w, r, address, http.StatusTemporaryRedirect)
	})
	return r
}

func SetShortURL(fURL string) string {

	for key, addr := range baseURL {
		if addr == fURL {
			return key
		}
	}
	for {
		key = keygen.RandomStr()
		_, err := baseURL[key]
		if !err {
			break
		}
	}
	var mutex sync.Mutex
	mutex.Lock()
	baseURL[key] = fURL
	mutex.Unlock()
	return key
}

func RetFullURL(key string) string {
	return baseURL[key]
}
