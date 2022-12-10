package handlers

import (
	"bytes"
	"encoding/json"
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
		{
			name:    "test #2",
			request: []byte(`/github.com/Yandex-Practicum/go-autotests`),
			want: want{
				contentType: "application/json; charset=utf-8",
				statusCode1: 201,
				statusCode2: 307,
			},
		},
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
			assert.Equal(t, tt.want.statusCode1, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			userResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
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
		})
	}
}
