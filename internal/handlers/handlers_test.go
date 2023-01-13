package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"shortURL/internal/config"
	"shortURL/internal/storage"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	Params := config.NewConfig()
	storage := storage.NewStorage(Params)
	r := NewRouter(Params, storage)

	l, err := net.Listen("tcp", Params.Server)
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(r)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()

	defer ts.Close()

	type want struct {
		contentType string
		statusCode1 int
		statusCode2 int
		statusCode3 int
	}

	type URLs struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	type MultiURL struct {
		CorrID    string `json:"correlation_id"`
		OriginURL string `json:"original_url,omitempty"`
		ShortURL  string `json:"short_url,omitempty"`
	}

	tests := []struct {
		name    string
		request []byte
		want    want
	}{
		{
			name:    "test #1",
			request: []byte(`/github.com/Yandex-Practicum/go-autotests`),
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode1: 201,
				statusCode2: 307,
				statusCode3: 409,
			},
		},
		{
			name:    "test #2",
			request: []byte(`/postgrespro.ru/docs/postgrespro/13/sql-syntax`),
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode1: 201,
				statusCode2: 307,
				statusCode3: 409,
			},
		},
	}
if Params.SaveFile == 2 {
	t.Run("Ping", func(t *testing.T) {
		request1, err := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
		require.NoError(t, err)
		result, err := http.DefaultClient.Do(request1)
		require.NoError(t, err)
		assert.Equal(t, 200, result.StatusCode)
		err = result.Body.Close()
		require.NoError(t, err)
	})
}


	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//POST
			var request1 *http.Request
			var err error
			switch tt.want.contentType {
			case "text/plain; charset=utf-8":
				request1, err = http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewReader(tt.request))
				require.NoError(t, err)
			case "application/json; charset=utf-8":
				var req PostURL
				req.GetURL = string(tt.request)
				reqBz, err := json.Marshal(req)
				if err != nil {
					panic(err)
				}
				request1, err = http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", bytes.NewReader(reqBz))
				require.NoError(t, err)
			}

			result, err := http.DefaultClient.Do(request1)
			require.NoError(t, err)
			var c http.Cookie
			for _, cookie := range result.Cookies() {
				if cookie.Name == "shortener" {
					c = *cookie
					break
				}
			}
			assert.Equal(t, tt.want.statusCode1, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			userResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			//POST & cookie
			var request3 *http.Request
			switch tt.want.contentType {
			case "text/plain; charset=utf-8":
				request3, err = http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewReader(tt.request))
				require.NoError(t, err)
			case "application/json; charset=utf-8":
				var req PostURL
				req.GetURL = string(tt.request)
				reqBz, err := json.Marshal(req)
				if err != nil {
					panic(err)
				}
				request3, err = http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", bytes.NewReader(reqBz))
				require.NoError(t, err)
			}
			request3.AddCookie(&c)
			result3, err := http.DefaultClient.Do(request3)
			require.NoError(t, err)
			assert.Equal(t, tt.want.statusCode3, result3.StatusCode)
			err = result3.Body.Close()
			require.NoError(t, err)

			//GET
			var request2 *http.Request
			switch tt.want.contentType {
			case "text/plain; charset=utf-8":
				request2, err = http.NewRequest(http.MethodGet, string(userResult), nil)
				require.NoError(t, err)
			case "application/json; charset=utf-8":
				var res PostURL
				err := json.Unmarshal(userResult, &res)
				if err != nil {
					panic(err)
				}
				request2, err = http.NewRequest(http.MethodGet, res.SetURL, nil)
				require.NoError(t, err)
			}

			result2, err := http.DefaultTransport.RoundTrip(request2)
			require.NoError(t, err)
			userResult2 := result2.Header.Get("Location")
			err = result2.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want.statusCode2, result2.StatusCode)
			assert.Equal(t, tt.request, []byte(userResult2))

			// GET urls
			var request4 *http.Request
			request4, err = http.NewRequest(http.MethodGet, ts.URL+"/api/user/urls", nil)
			require.NoError(t, err)
			result4, err := http.DefaultClient.Do(request4)
			require.NoError(t, err)
			assert.Equal(t, 204, result4.StatusCode)
			err = result4.Body.Close()
			require.NoError(t, err)

			// GET urls & cookie
			var request5 *http.Request
			request5, err = http.NewRequest(http.MethodGet, ts.URL+"/api/user/urls", nil)
			require.NoError(t, err)
			request5.AddCookie(&c)
			result5, err := http.DefaultClient.Do(request5)
			require.NoError(t, err)
			assert.Equal(t, 200, result5.StatusCode)
			userResult5, err := io.ReadAll(result5.Body)
			require.NoError(t, err)
			err = result5.Body.Close()
			require.NoError(t, err)
			var AllURLs = make([]URLs, 0)
			err = json.Unmarshal(userResult5, &AllURLs)
			require.NoError(t, err)
			for _, u := range AllURLs {
				log.Println(u.OriginalURL, u.ShortURL)
			}
		})
	}

	t.Run("MultiURL", func(t *testing.T) {
		Multi := []struct {
			mult MultiURL
		}{
			{
				mult: MultiURL{
					CorrID:    "abc123",
					OriginURL: "/github.com/Yandex-Practicum/go-autotests",
				},
			},
			{
				mult: MultiURL{
					CorrID:    "def456",
					OriginURL: "/postgrespro.ru/docs/postgrespro/13/sql-syntax",
				},
			},
		}
		RMultiURLsBZ, err := json.Marshal(Multi)
		if err != nil {
			log.Fatal(err)
		}

		request1, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", bytes.NewReader(RMultiURLsBZ))
		require.NoError(t, err)
		result, err := http.DefaultClient.Do(request1)
		require.NoError(t, err)
		assert.Equal(t, 201, result.StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", result.Header.Get("Content-Type"))
		userResult1, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		err = result.Body.Close()
		require.NoError(t, err)
		Multis := make([]MultiURL, 2)
		err = json.Unmarshal(userResult1, &Multis)
		require.NoError(t, err)
		for _, u := range Multis {
			log.Println(u.CorrID, u.ShortURL)
		}
		log.Println("Done")
	})

}
