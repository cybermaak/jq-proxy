package logging

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(t *testing.T) *Logger {
	t.Helper()
	logger, err := NewLogger("error")
	require.NoError(t, err)
	return logger
}

func TestRequestLoggingMiddleware_SetsRequestIDInContext(t *testing.T) {
	logger := newTestLogger(t)

	var capturedID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestLoggingMiddleware(logger)(next)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.NotEmpty(t, capturedID, "request ID should be injected into context")
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequestLoggingMiddleware_CapturesStatusCode(t *testing.T) {
	logger := newTestLogger(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := RequestLoggingMiddleware(logger)(next)
	req := httptest.NewRequest(http.MethodPost, "/proxy/svc/items", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestRequestLoggingMiddleware_CapturesBytesWritten(t *testing.T) {
	logger := newTestLogger(t)
	body := []byte(`{"ok":true}`)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})

	handler := RequestLoggingMiddleware(logger)(next)
	req := httptest.NewRequest(http.MethodGet, "/proxy/svc/items", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, body, rr.Body.Bytes())
}

func TestRequestLoggingMiddleware_DefaultStatusIsOK(t *testing.T) {
	logger := newTestLogger(t)

	// Handler writes body without calling WriteHeader explicitly — default should be 200.
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})

	handler := RequestLoggingMiddleware(logger)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// responseWriter tests

func TestResponseWriter_WriteHeader_CapturesCode(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusNotFound)

	assert.Equal(t, http.StatusNotFound, rw.statusCode)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestResponseWriter_Write_AccumulatesBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, statusCode: http.StatusOK}

	chunk1 := []byte("hello")
	chunk2 := []byte(" world")

	n1, err := rw.Write(chunk1)
	require.NoError(t, err)
	n2, err := rw.Write(chunk2)
	require.NoError(t, err)

	assert.Equal(t, len(chunk1), n1)
	assert.Equal(t, len(chunk2), n2)
	assert.Equal(t, len(chunk1)+len(chunk2), rw.bytesWritten)
	assert.Equal(t, append(chunk1, chunk2...), rec.Body.Bytes())
}

// getClientIP tests

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")

	ip := getClientIP(req)

	assert.Equal(t, "203.0.113.1, 10.0.0.1", ip)
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.5")

	ip := getClientIP(req)

	assert.Equal(t, "198.51.100.5", ip)
}

func TestGetClientIP_FallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// No proxy headers — should fall back to the connection address.
	req.RemoteAddr = "192.0.2.42:54321"

	ip := getClientIP(req)

	assert.Equal(t, "192.0.2.42:54321", ip)
}

func TestGetClientIP_XForwardedForTakesPrecedenceOverXRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.Header.Set("X-Real-IP", "198.51.100.5")

	ip := getClientIP(req)

	// X-Forwarded-For wins when both are set.
	assert.Equal(t, "203.0.113.1", ip)
}

func TestRequestLoggingMiddleware_UniqueRequestIDPerRequest(t *testing.T) {
	logger := newTestLogger(t)

	ids := make([]string, 3)
	for i := range ids {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ids[i] = GetRequestID(r.Context())
		})
		handler := RequestLoggingMiddleware(logger)(next)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	assert.NotEqual(t, ids[0], ids[1])
	assert.NotEqual(t, ids[1], ids[2])
}

func TestRequestLoggingMiddleware_PassesBodyThrough(t *testing.T) {
	logger := newTestLogger(t)
	reqBody := []byte(`{"method":"GET"}`)

	var receivedBody []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		receivedBody = buf.Bytes()
	})

	handler := RequestLoggingMiddleware(logger)(next)
	req := httptest.NewRequest(http.MethodPost, "/proxy/svc", bytes.NewReader(reqBody))
	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, reqBody, receivedBody)
}
