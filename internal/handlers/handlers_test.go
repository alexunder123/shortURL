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
	"shortURL/internal/router"
	"shortURL/internal/storage"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type urls struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type want struct {
	contentType string
	statusCode1 int
	statusCode2 int
	statusCode3 int
}

type test struct {
	name    string
	request []byte
	want    want
}

func TestRouter(t *testing.T) {
	params := config.NewConfig()
	storage := storage.NewStorage(params)
	r := router.NewRouter(params, storage)
	h := NewHandler(r)

	l, err := net.Listen("tcp", params.Server)
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(h)
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()

	defer ts.Close()

	tests := []test{
		{
			name:    "test #1 text",
			request: []byte(`/github.com/Yandex-Practicum/go-autotests`),
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode1: 201,
				statusCode2: 307,
				statusCode3: 409,
			},
		},
		{
			name:    "test #2 json",
			request: []byte(`/postgrespro.ru/docs/postgrespro/13/sql-syntax`),
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode1: 201,
				statusCode2: 307,
				statusCode3: 409,
			},
		},
	}
	if params.SavePlace == config.SaveSQL {
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
			userResult, c := postURL(ts, t, tt)

			//GET
			getURL(ts, t, tt, userResult)

			// GET URLs
			getUserURLs(ts, t, tt, userResult, c)
		})
	}
	multiURL(ts, t)
	log.Println("Done")

}

func postURL(ts *httptest.Server, t *testing.T, tt test) (userResult []byte, c http.Cookie) {
	//POST
	var request1 *http.Request
	var err error
	switch tt.want.contentType {
	case "text/plain; charset=utf-8":
		request1, err = http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewReader(tt.request))
		require.NoError(t, err)
	case "application/json; charset=utf-8":
		var req router.PostURL
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
	for _, cookie := range result.Cookies() {
		if cookie.Name == "shortener" {
			c = *cookie
			break
		}
	}
	assert.Equal(t, tt.want.statusCode1, result.StatusCode)
	assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
	userResult, err = io.ReadAll(result.Body)
	require.NoError(t, err)
	err = result.Body.Close()
	require.NoError(t, err)

	//POST & cookie
	request1.AddCookie(&c)
	result3, err := http.DefaultClient.Do(request1)
	require.NoError(t, err)
	assert.Equal(t, tt.want.statusCode3, result3.StatusCode)
	err = result3.Body.Close()
	require.NoError(t, err)

	return
}

func getURL(ts *httptest.Server, t *testing.T, tt test, userResult []byte) {
	var request2 *http.Request
	var err error
	switch tt.want.contentType {
	case "text/plain; charset=utf-8":
		request2, err = http.NewRequest(http.MethodGet, string(userResult), nil)
		require.NoError(t, err)
	case "application/json; charset=utf-8":
		var res router.PostURL
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
}

func getUserURLs(ts *httptest.Server, t *testing.T, tt test, userResult []byte, c http.Cookie) {
	// GET urls
	var request4 *http.Request
	request4, err := http.NewRequest(http.MethodGet, ts.URL+"/api/user/urls", nil)
	require.NoError(t, err)
	result4, err := http.DefaultClient.Do(request4)
	require.NoError(t, err)
	assert.Equal(t, 204, result4.StatusCode)
	err = result4.Body.Close()
	require.NoError(t, err)

	// GET urls & cookie
	request4.AddCookie(&c)
	result5, err := http.DefaultClient.Do(request4)
	require.NoError(t, err)
	assert.Equal(t, 200, result5.StatusCode)
	userResult5, err := io.ReadAll(result5.Body)
	require.NoError(t, err)
	err = result5.Body.Close()
	require.NoError(t, err)
	var AllURLs = make([]urls, 0)
	err = json.Unmarshal(userResult5, &AllURLs)
	require.NoError(t, err)
	for _, u := range AllURLs {
		log.Println(u.OriginalURL, u.ShortURL)
	}
}

func multiURL(ts *httptest.Server, t *testing.T) {
	type multiURL struct {
		CorrID    string `json:"correlation_id"`
		OriginURL string `json:"original_url,omitempty"`
		ShortURL  string `json:"short_url,omitempty"`
	}
	t.Run("MultiURL", func(t *testing.T) {
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
			log.Fatal(err)
		}

		request1, err := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", bytes.NewReader(multiURLsBZ))
		require.NoError(t, err)
		result, err := http.DefaultClient.Do(request1)
		require.NoError(t, err)
		assert.Equal(t, 201, result.StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", result.Header.Get("Content-Type"))
		userResult1, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		err = result.Body.Close()
		require.NoError(t, err)
		multis := make([]multiURL, 2)
		err = json.Unmarshal(userResult1, &multis)
		require.NoError(t, err)
		for _, u := range multis {
			log.Println(u.CorrID, u.ShortURL)
		}
	})
}
