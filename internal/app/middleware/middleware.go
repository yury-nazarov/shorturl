package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

//contentType := ["application/javascript", "application/json", "text/css", "text/html", "text/plain", "text/xml"]

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// HTTPResponseCompressor - отправляет сжатый gzip HTTP Response,
//							если от клиента пришел заголовок: "Accept-Encoding: gzip"
func HTTPResponseCompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		// Если gzip не поддерживается, передаем управление дальше без изменений
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip"){
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)

	})
}


// HTTPRequestDecompressor - распаковывает сжаты gzip HTTP Request Body.
func HTTPRequestDecompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		// Если запрос не сжат с помощью gzip, передаем управление дальше без изменений
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer gz.Close()
		data, err := io.ReadAll(gz)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body := ioutil.NopCloser(bytes.NewBuffer(data))
		r.Body = body
		next.ServeHTTP(w, r)
	})
}


