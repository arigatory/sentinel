package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arigatory/sentinel/internal/hash"
)

const testKey = "supersecret"

// echoHandler отвечает фиксированным телом и запоминает, что до него дошло.
func echoHandler(called *bool, gotBody *[]byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		body, _ := io.ReadAll(r.Body)
		*gotBody = body

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
}

func TestHashMiddleware(t *testing.T) {
	body := `{"id":"Test","type":"gauge","value":1.5}`

	tests := []struct {
		name       string
		key        string
		header     string // значение заголовка HashSHA256; "-" означает "не ставить"
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "valid signature",
			key:        testKey,
			header:     hash.Sign([]byte(body), testKey),
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "invalid signature",
			key:        testKey,
			header:     hash.Sign([]byte(body), "wrongkey"),
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
		{
			name:       "malformed signature",
			key:        testKey,
			header:     "not-a-hex-value",
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
		{
			name:       "no header with key configured",
			key:        testKey,
			header:     "-",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "legacy none header",
			key:        testKey,
			header:     "none",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
		{
			name:       "empty key disables middleware",
			key:        "",
			header:     "garbage",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var called bool
			var gotBody []byte
			h := HashMiddleware(tt.key)(echoHandler(&called, &gotBody))

			request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(body))
			if tt.header != "-" {
				request.Header.Set(hash.Header, tt.header)
			}
			w := httptest.NewRecorder()

			// Act
			h.ServeHTTP(w, request)

			// Assert
			result := w.Result()
			defer result.Body.Close()

			if result.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, result.StatusCode)
			}
			if called != tt.wantCalled {
				t.Errorf("Expected handler called=%v, got %v", tt.wantCalled, called)
			}
			if tt.wantCalled && string(gotBody) != body {
				t.Errorf("Expected handler to read body %q, got %q", body, string(gotBody))
			}
		})
	}
}

func TestHashMiddlewareSignsResponse(t *testing.T) {
	// Arrange
	var called bool
	var gotBody []byte
	h := HashMiddleware(testKey)(echoHandler(&called, &gotBody))

	request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()

	// Act
	h.ServeHTTP(w, request)

	// Assert
	result := w.Result()
	defer result.Body.Close()

	respBody, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	got := result.Header.Get(hash.Header)
	if got == "" {
		t.Fatalf("Expected response header %s to be set", hash.Header)
	}
	if !hash.Valid(respBody, testKey, got) {
		t.Errorf("Response signature %q does not match body %q", got, string(respBody))
	}
	if ct := result.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type to survive the middleware, got %q", ct)
	}
}

func TestHashMiddlewareNoResponseHeaderWithoutKey(t *testing.T) {
	var called bool
	var gotBody []byte
	h := HashMiddleware("")(echoHandler(&called, &gotBody))

	request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, request)

	result := w.Result()
	defer result.Body.Close()

	if got := result.Header.Get(hash.Header); got != "" {
		t.Errorf("Expected no %s header without a key, got %q", hash.Header, got)
	}
}

// TestHashMiddlewareWithGzip воспроизводит порядок middleware из cmd/server/main.go:
// Gzip распаковывает тело, Hash проверяет уже распакованные байты.
func TestHashMiddlewareWithGzip(t *testing.T) {
	// Arrange
	body := `{"id":"Test","type":"gauge","value":1.5}`

	var compressed bytes.Buffer
	gw := gzip.NewWriter(&compressed)
	if _, err := gw.Write([]byte(body)); err != nil {
		t.Fatalf("Failed to compress body: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	var called bool
	var gotBody []byte
	h := GzipMiddleware(HashMiddleware(testKey)(echoHandler(&called, &gotBody)))

	request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(compressed.Bytes()))
	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Accept-Encoding", "gzip")
	// подпись считается от несжатого тела
	request.Header.Set(hash.Header, hash.Sign([]byte(body), testKey))
	w := httptest.NewRecorder()

	// Act
	h.ServeHTTP(w, request)

	// Assert
	result := w.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, result.StatusCode)
	}
	if !called {
		t.Fatal("Expected handler to be called")
	}
	if string(gotBody) != body {
		t.Errorf("Expected handler to read %q, got %q", body, string(gotBody))
	}
	if enc := result.Header.Get("Content-Encoding"); enc != "gzip" {
		t.Errorf("Expected response to stay gzip-compressed, got Content-Encoding %q", enc)
	}
	if result.Header.Get(hash.Header) == "" {
		t.Errorf("Expected response to carry the %s header", hash.Header)
	}
}
