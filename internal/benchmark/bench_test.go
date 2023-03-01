package bencmark

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"shortURL/internal/config"
	"shortURL/internal/handler"
	"shortURL/internal/midware"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
)

type postURLs struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}

type multiURL struct {
	CorrID    string `json:"correlation_id"`
	OriginURL string `json:"original_url,omitempty"`
	ShortURL  string `json:"short_url,omitempty"`
}

func BenchmarkRouter(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
	cnfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("NewConfig read environment error")
	}
	storage := storage.NewStorage(cnfg)
	deletingWorker := worker.NewWorker()
	handlers := handler.NewHandler(cnfg, storage, deletingWorker)
	router := NewRouter(handlers)
	deletingWorker.Run(storage, cnfg.DeletingBufferSize, cnfg.DeletingBufferTimeout)
	l, err := net.Listen("tcp", cnfg.ServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("net.Listen error")
	}

	ts := httptest.NewUnstartedServer(router)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()

	defer ts.Close()
	const qTests = 50

	var request *http.Request
	var cookie http.Cookie
	var userResult []byte

	var res postURLs

	multi := []multiURL{
		{
			CorrID:    "abc123",
			OriginURL: "/github.com/Yandex-Practicum/go-autotests",
		},
		{
			CorrID:    "def456",
			OriginURL: "/postgrespro.ru/docs/postgrespro/13/sql-syntax",
		},
	}

	multiURLsBZ, err := json.Marshal(multi)
	if err != nil {
		log.Fatal().Err(err).Msg("json.Marshal error")
	}

	tests := []struct {
		request []byte
	}{
		{
			request: []byte(`/tapoueh.org/blog/2018/07/batch-updates-and-concurrency`),
		},
		{
			request: []byte(`/pkg.go.dev/bytes`),
		},
		{
			request: []byte(`/pkg.go.dev/strings`),
		},
		{
			request: []byte(`/pkg.go.dev/database/sql`),
		},
		{
			request: []byte(`https://practicum.yandex.ru/learn/go-advanced/courses`),
		},
		{
			request: []byte(`/pkg.go.dev/net/url`),
		},
		{
			request: []byte(`/pkg.go.dev/builtin`),
		},
		{
			request: []byte(`/www.youtube.com`),
		},
	}

	b.ResetTimer()

	// for {
	b.Run("postURL text", func(b *testing.B) {
		for i := 0; i < qTests; i++ {
			request, err = http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewReader([]byte(`/github.com/Yandex-Practicum/go-autotests`)))
			if err != nil {
				log.Fatal().Err(err).Msg("http.NewRequest error")
			}
			result, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			for _, cook := range result.Cookies() {
				if cook.Name == "shortener" {
					cookie = *cook
					break
				}
			}
			userResult, err = io.ReadAll(result.Body)
			if err != nil {
				log.Fatal().Err(err).Msg("io.ReadAll error")
			}
			result.Body.Close()

			//POST & cookie
			request.AddCookie(&cookie)
			result1, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("io.ReadAll error")
			}
			result1.Body.Close()
		}
	})

	b.Run("getURL text", func(b *testing.B) {
		for i := 0; i < qTests; i++ {
			request, err = http.NewRequest(http.MethodGet, string(userResult), nil)
			if err != nil {
				log.Fatal().Err(err).Msg("http.NewRequest error")
			}
			result, err := http.DefaultTransport.RoundTrip(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			result.Body.Close()
		}
	})

	b.Run("postURL JSON", func(b *testing.B) {
		var req postURLs
		req.GetURL = "/habr.com/ru/users/triumphpc/posts/"
		reqBz, err := json.Marshal(req)
		if err != nil {
			log.Fatal().Err(err).Msg("json.Marshal error")
		}
		for i := 0; i < qTests; i++ {
			request, err = http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", bytes.NewReader(reqBz))
			if err != nil {
				log.Fatal().Err(err).Msg("http.NewRequest error")
			}
			result, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			for _, cook := range result.Cookies() {
				if cook.Name == "shortener" {
					cookie = *cook
					break
				}
			}
			userResult, err = io.ReadAll(result.Body)
			if err != nil {
				log.Fatal().Err(err).Msg("io.ReadAll error")
			}
			result.Body.Close()

			//POST & cookie
			request.AddCookie(&cookie)
			result1, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("io.ReadAll error")
			}
			result1.Body.Close()
		}
	})

	b.Run("getURL JSON", func(b *testing.B) {
		for i := 0; i < qTests; i++ {
			err := json.Unmarshal(userResult, &res)
			if err != nil {
				log.Fatal().Err(err).Msg("json.Unmarshal error")
			}
			request, err = http.NewRequest(http.MethodGet, res.SetURL, nil)
			if err != nil {
				log.Fatal().Err(err).Msg("http.NewRequest error")
			}
			result, err := http.DefaultTransport.RoundTrip(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			result.Body.Close()
		}
	})

	b.Run("getUserURLs", func(b *testing.B) {
		for i := 0; i < qTests; i++ {
			request, err = http.NewRequest(http.MethodGet, ts.URL+"/api/user/urls", nil)
			if err != nil {
				log.Fatal().Err(err).Msg("http.NewRequest error")
			}
			result, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			result.Body.Close()

			// GET urls & cookie
			request.AddCookie(&cookie)
			result1, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			result1.Body.Close()
		}
	})

	b.Run("postBatchURLs", func(b *testing.B) {
		for i := 0; i < qTests; i++ {
			request, err = http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", bytes.NewReader(multiURLsBZ))
			if err != nil {
				log.Fatal().Err(err).Msg("http.NewRequest error")
			}
			result, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			result.Body.Close()
		}
	})

	b.Run("postURLsToDelete", func(b *testing.B) {
		for i := 0; i < qTests; i++ {
			//GET cookie
			request, err = http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
			if err != nil {
				log.Fatal().Err(err).Msg("http.NewRequest error")
			}
			result, err := http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			result.Body.Close()

			for _, cook := range result.Cookies() {
				if cook.Name == "shortener" {
					cookie = *cook
					break
				}
			}

			//POST URLs with cookie
			results := make([]string, 0, 8)

			deletes := make([]string, 0, 8)
			for _, tt := range tests {
				request, err = http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewReader(tt.request))
				if err != nil {
					log.Fatal().Err(err).Msg("http.NewRequest error")
				}
				request.AddCookie(&cookie)
				result, err := http.DefaultClient.Do(request)
				if err != nil {
					log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
				}
				userResult, err := io.ReadAll(result.Body)
				if err != nil {
					log.Fatal().Err(err).Msg("io.ReadAll error")
				}
				result.Body.Close()

				results = append(results, string(userResult))
				j := bytes.LastIndex(userResult, []byte(`/`))
				if j == -1 {
					continue
				}
				res := userResult[j+1:]
				deletes = append(deletes, string(res))
			}

			deletesBZ, err := json.Marshal(deletes)
			if err != nil {
				log.Fatal().Err(err).Msg("json.Marshal error")
			}
			// DELETE URLs
			request, err = http.NewRequest(http.MethodDelete, ts.URL+"/api/user/urls", bytes.NewReader([]byte(deletesBZ)))
			if err != nil {
				log.Fatal().Err(err).Msg("http.NewRequest error")
			}
			request.AddCookie(&cookie)
			result, err = http.DefaultClient.Do(request)
			if err != nil {
				log.Fatal().Err(err).Msg("http.DefaultClient.Do error")
			}
			result.Body.Close()

			//GET deleted URLs
			b.StopTimer()
			time.Sleep(10 * time.Millisecond)
			b.StartTimer()
			for _, res := range results {
				request, err = http.NewRequest(http.MethodGet, res, nil)
				if err != nil {
					log.Fatal().Err(err).Msg("http.NewRequest error")
				}
				request.AddCookie(&cookie)
				result, err = http.DefaultTransport.RoundTrip(request)
				if err != nil {
					log.Fatal().Err(err).Msg("http.DefaultTransport.RoundTrip error")
				}
				result.Body.Close()
			}
		}
	})

	// }

}

func NewRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(midware.Decompress)
	r.Use(midware.Cookies)

	r.Post("/api/shorten/batch", h.BatchNewEtriesPost)
	r.Post("/api/shorten", h.ShortenPost)
	r.Post("/", h.URLPost)

	r.Get("/api/user/urls", h.URLsGet)
	r.Get("/{id}", h.IDGet)
	r.Get("/ping", h.PingGet)

	r.Delete("/api/user/urls", h.URLsDelete)

	return r
}
