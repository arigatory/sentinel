package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

var compressibleTypes = []string{
	"application/json",
	"text/html",
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
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

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = gr
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		lw := &lazyGzipWriter{ResponseWriter: w}
		defer lw.close()

		next.ServeHTTP(lw, r)
	})
}

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
