package midware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// gzipWriter добавляем обертку к интерфейсу, для замены штатного метода.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write добавляем свой метод.
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Decompress middleware для сжатия и расшифровки передаваемых данных.
func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			decompressed := gz
			defer gz.Close()
			r.Body = io.NopCloser(decompressed)
			next.ServeHTTP(w, r)
			return
		} else if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			lenght, _ := strconv.Atoi(w.Header().Get("Content-Length"))
			if lenght < 1400 { //1400 (1.4kB) минимальный объем данных, меньше которого нет смысла сжимать
				next.ServeHTTP(w, r)
				return
			}
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
			return
		} else {
			next.ServeHTTP(w, r)
			return
		}
	})
}
