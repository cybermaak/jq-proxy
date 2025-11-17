// Package logging provides structured logging and metrics collection functionality.
package logging

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// RequestLoggingMiddleware creates middleware for request logging with tracing
func RequestLoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate request ID
			requestID := GenerateRequestID()
			ctx := WithRequestIDContext(r.Context(), requestID)
			r = r.WithContext(ctx)

			// Create response writer wrapper
			wrapper := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Record start time
			startTime := time.Now()

			// Log incoming request
			logger.WithRequestID(ctx).WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"query":      r.URL.RawQuery,
				"user_agent": r.UserAgent(),
				"remote_ip":  getClientIP(r),
			}).Info("Request started")

			// Process request
			next.ServeHTTP(wrapper, r)

			// Calculate duration
			duration := time.Since(startTime)

			// Log completed request
			logger.WithRequestID(ctx).WithFields(logrus.Fields{
				"method":        r.Method,
				"path":          r.URL.Path,
				"status_code":   wrapper.statusCode,
				"duration_ms":   duration.Milliseconds(),
				"response_size": wrapper.bytesWritten,
			}).Info("Request completed")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
