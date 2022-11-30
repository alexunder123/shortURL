package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	r := NewRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		contentType string
		statusCode1 int
		statusCode2 int
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//POST
			request1, err := http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewReader(tt.request))
			require.NoError(t, err)

			result, err := http.DefaultClient.Do(request1)
			require.NoError(t, err)
			// w := httptest.NewRecorder()
			// h := http.HandlerFunc(ShortenerURL)
			// h(w, request)
			// result := w.Result()
			assert.Equal(t, tt.want.statusCode1, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			userResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			
			//GET
			request2, err := http.NewRequest(http.MethodGet, string(userResult), nil)
			require.NoError(t, err)

			result2, err := http.DefaultTransport.RoundTrip(request2)
			require.NoError(t, err)
			// w2 := httptest.NewRecorder()
			// h2 := http.HandlerFunc(ShortenerURL)
			// h2(w2, request2)
			// result2 := w2.Result()
			userResult2 := result2.Header.Get("Location")
			// err = result2.Body.Close()
			// require.NoError(t, err)
			assert.Equal(t, tt.want.statusCode2, result2.StatusCode)
			assert.Equal(t, tt.request, []byte(userResult2))
		})
	}
}
