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

type gzipBodyWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipBodyWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// HTTPResponseCompressor - от клиента пришел заголовок: "Accept-Encoding: gzip"
//							(Данные от клиента передал одним из текстовых форматов)
//							вернет сжатый gzip HTTP Response Body.
func HTTPResponseCompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если gzip не поддерживается, передаем управление дальше без изменений
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
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

		// устанавливаем заголовки
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")

		next.ServeHTTP(gzipBodyWriter{
			ResponseWriter: w,
			Writer:         gz,
		}, r)

	})
}

// HTTPRequestDecompressor - от клиента пришел заголовок: "Content-Encoding: gzip"
//							 (Данные от клиента сжаты в формате gzip!)
//  						 распаковывает сжатый gzip HTTP Request Body.
//							 Добавляет в БД текстовую ссылку
func HTTPRequestDecompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если запрос не сжат с помощью gzip, передаем управление дальше без изменений
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
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
		// Записываем распакованый в текстовый формат body и передаем управление дальше
		body := ioutil.NopCloser(bytes.NewBuffer(data))
		r.Body = body
		next.ServeHTTP(w, r)
	})
}
