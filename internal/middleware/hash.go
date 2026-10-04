package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/arigatory/sentinel/internal/hash"
)

// noneHash — легаси-значение заголовка подписи, которое встречается у некоторых
// клиентов и означает "подпись не рассчитывалась".
const noneHash = "none"

// hashResponseWriter буферизует тело ответа, чтобы подпись можно было посчитать
// до того, как байты уйдут клиенту. Header() намеренно не переопределяется:
// заголовки, выставленные хендлером, попадают сразу в исходный ResponseWriter.
type hashResponseWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	statusCode  int
	wroteHeader bool
}

func (hw *hashResponseWriter) WriteHeader(statusCode int) {
	if hw.wroteHeader {
		return
	}
	hw.statusCode = statusCode
	hw.wroteHeader = true
}

func (hw *hashResponseWriter) Write(b []byte) (int, error) {
	if !hw.wroteHeader {
		hw.WriteHeader(http.StatusOK)
	}
	return hw.buf.Write(b)
}

// HashMiddleware проверяет подпись входящего запроса и подписывает ответ.
//
// Пустой key полностью отключает обработку. Запросы без заголовка подписи
// пропускаются без проверки — отвергаются только те, где заголовок есть,
// но не совпадает с расчётным значением.
//
// Middleware должен стоять после GzipMiddleware, чтобы работать с распакованным телом.
func HashMiddleware(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get(hash.Header); got != "" && got != noneHash {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read request body", http.StatusBadRequest)
					return
				}
				r.Body.Close()
				r.Body = io.NopCloser(bytes.NewReader(body))

				if !hash.Valid(body, key, got) {
					http.Error(w, "Invalid hash", http.StatusBadRequest)
					return
				}
			}

			hw := &hashResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(hw, r)

			body := hw.buf.Bytes()
			w.Header().Set(hash.Header, hash.Sign(body, key))
			w.WriteHeader(hw.statusCode)
			if len(body) > 0 {
				w.Write(body)
			}
		})
	}
}
