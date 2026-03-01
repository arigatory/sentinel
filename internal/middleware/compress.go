package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressibleTypes — типы контента, которые нужно сжимать.
var compressibleTypes = []string{
	"application/json",
	"text/html",
}

// gzipResponseWriter оборачивает http.ResponseWriter и пишет сжатые данные.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer // либо gzip.Writer, либо оригинальный ResponseWriter
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

func isCompressible(contentType string) bool {
	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

// GzipMiddleware сжимает ответ при наличии Accept-Encoding: gzip
// и распаковывает входящий запрос при наличии Content-Encoding: gzip.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Декомпрессия входящего запроса
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = gr
		}

		// Компрессия ответа, если клиент поддерживает gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Используем lazy-writer: сжимаем только compressible типы.
		// Оборачиваем ResponseWriter в перехватчик, который при первом
		// вызове WriteHeader/Write решает, нужно ли сжимать.
		lw := &lazyGzipWriter{ResponseWriter: w}
		defer lw.close()

		next.ServeHTTP(lw, r)
	})
}

// lazyGzipWriter откладывает решение о сжатии до момента, когда известен Content-Type.
type lazyGzipWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	initialized bool
	compress    bool
}

func (l *lazyGzipWriter) WriteHeader(code int) {
	l.init()
	l.ResponseWriter.WriteHeader(code)
}

func (l *lazyGzipWriter) Write(b []byte) (int, error) {
	l.init()
	if l.compress {
		return l.gz.Write(b)
	}
	return l.ResponseWriter.Write(b)
}

func (l *lazyGzipWriter) init() {
	if l.initialized {
		return
	}
	l.initialized = true
	ct := l.ResponseWriter.Header().Get("Content-Type")
	if isCompressible(ct) {
		l.compress = true
		l.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		l.ResponseWriter.Header().Del("Content-Length")
		l.gz = gzip.NewWriter(l.ResponseWriter)
	}
}

func (l *lazyGzipWriter) close() {
	if l.gz != nil {
		l.gz.Close()
	}
}
