package router

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shortURL/internal/config"
	"shortURL/internal/handler"
	"shortURL/internal/storage"
	"shortURL/internal/worker"
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

type postURLs struct {
	GetURL string `json:"url,omitempty"`
	SetURL string `json:"result,omitempty"`
}

func TestRouter(t *testing.T) {
	// os.Setenv("DATABASE_DSN", "postgres://postgres:1@localhost:5432/postgres?sslmode=disable")
	os.Setenv("TRUSTED_SUBNET", "192.168.11.0/24")
	cnfg, err := config.NewConfig()
	require.NoError(t, err)
	storage := storage.NewStorage(cnfg)
	deletingWorker := worker.NewWorker()
	handlers := handler.NewHandler(cnfg, storage, deletingWorker)
	router := NewRouter(handlers)
	deletingWorker.Run(storage, cnfg.DeletingBufferSize, cnfg.DeletingBufferTimeout)
	listener, err := net.Listen("tcp", cnfg.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	testServer := httptest.NewUnstartedServer(router)
	testServer.Listener.Close()
	testServer.Listener = listener
	testServer.Start()

	defer testServer.Close()

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
	if cnfg.SavePlace == config.SaveSQL {
		t.Run("Ping", func(t *testing.T) {
			request1, err := http.NewRequest(http.MethodGet, testServer.URL+"/ping", nil)
			require.NoError(t, err)
			result, err := http.DefaultClient.Do(request1)
			require.NoError(t, err)
			assert.Equal(t, 200, result.StatusCode)
			err = result.Body.Close()
			require.NoError(t, err)
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			//POST
			userResult, c := postURL(testServer, t, test)

			//GET
			getURL(testServer, t, test, userResult)

			// GET URLs
			getUserURLs(testServer, t, test, userResult, c)
		})
	}
	multiURL(testServer, t)

	DeletedURL(testServer, t)

	getStats(testServer, t)

	deletingWorker.Stop()
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
		var req postURLs
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
		var res postURLs
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
	})

}

func DeletedURL(ts *httptest.Server, t *testing.T) {
	t.Run("DeletedURL", func(t *testing.T) {
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
		//GET cookie
		request0, err := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
		require.NoError(t, err)
		result, err := http.DefaultClient.Do(request0)
		require.NoError(t, err)
		err = result.Body.Close()
		require.NoError(t, err)
		var c http.Cookie
		for _, cookie := range result.Cookies() {
			if cookie.Name == "shortener" {
				c = *cookie
				break
			}
		}
		//POST URLs with cookie
		results := make([]string, 0, 8)

		deletes := make([]string, 0, 8)
		for _, tt := range tests {
			request1, err := http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewReader(tt.request))
			require.NoError(t, err)
			request1.AddCookie(&c)
			result, err := http.DefaultClient.Do(request1)
			require.NoError(t, err)
			assert.Equal(t, 201, result.StatusCode)
			userResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
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
			log.Fatal(err)
		}
		// DELETE URLs
		request2, err := http.NewRequest(http.MethodDelete, ts.URL+"/api/user/urls", bytes.NewReader([]byte(deletesBZ)))
		require.NoError(t, err)
		request2.AddCookie(&c)
		result, err = http.DefaultClient.Do(request2)
		require.NoError(t, err)
		assert.Equal(t, 202, result.StatusCode)
		err = result.Body.Close()
		require.NoError(t, err)

		//GET deleted URLs
		time.Sleep(300 * time.Millisecond)
		for _, res := range results {
			request3, err := http.NewRequest(http.MethodGet, res, nil)
			require.NoError(t, err)
			request3.AddCookie(&c)
			result, err := http.DefaultTransport.RoundTrip(request3)
			require.NoError(t, err)
			assert.Equal(t, 410, result.StatusCode)
			err = result.Body.Close()
			require.NoError(t, err)
		}
	})
}

func getStats(ts *httptest.Server, t *testing.T) {
	request, err := http.NewRequest(http.MethodGet, ts.URL+"/api/internal/stats", nil)
	require.NoError(t, err)
	result, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	err = result.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, result.StatusCode)

	request.Header.Add("X-Real-IP", "192.168.11.22")
	result, err = http.DefaultClient.Do(request)
	require.NoError(t, err)
	err = result.Body.Close()
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.StatusCode)
}
